/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import { inject as service } from '@ember/service';
import RepositoryService from 'sabo-ui/services/repository';
import { getOwner } from '@ember/application';
import { set } from '@ember/object';
import dataSource from 'sabo-ui/decorators/data-source';

const modelName = 'oidc-provider';
const OAUTH_PROVIDER_NAME = 'oidc-with-url';
export default class OidcProviderService extends RepositoryService {
  @service('torii') manager;
  @service('settings') settings;
  @service('env') env;

  init() {
    super.init(...arguments);
    this.provider = getOwner(this).lookup(`torii-provider:${OAUTH_PROVIDER_NAME}`);
  }

  getModelName() {
    return modelName;
  }

  @dataSource('/:partition/:ns/:dc/oidc/providers')
  async findAllByDatacenter() {
    const res = await super.findAllByDatacenter(...arguments);
    if (res.length === 0) {
      const err = new Error('Not found');
      err.statusCode = 404;
      this.store.adapterFor(this.getModelName()).error(err);
      return;
    }
    return res;
  }

  @dataSource('/:partition/:ns/:dc/oidc/provider/:id')
  async findBySlug(params) {
    // This addition is mainly due to ember-data book-keeping This is one of
    // the only places where Consul w/namespaces enabled doesn't return a
    // response with a Namespace property, but in order to keep ember-data
    // id's happy we need to fake one. Usually when we make a request to consul
    // with an empty `ns=` Consul will use the namespace that is assigned to
    // the token, and when we get the response we can pick that back off the
    // responses `Namespace` property. As we don't receive a `Namespace`
    // property here, we have to figure this out ourselves. But we also want
    // to make this completely invisible to 'the application engineer/a
    // template engineer'. This feels like the best place/way to do it as we
    // are already in a asynchronous method, and we avoid adding extra 'just
    // for us' parameters to the query object. There is a chance that as we
    // are discovering the tokens default namespace on the frontend and
    // assigning that to the ns query param, the token default namespace 'may'
    // have changed by the time the request hits the backend. As this is
    // extremely unlikely and in the scheme of things not a big deal, we
    // decided that doing this here is ok and avoids doing this in a more
    // complicated manner.
    const token = (await this.settings.findBySlug('token')) || {};
    const res = await super.findBySlug({
      ns: params.ns || token.Namespace || 'default',
      partition: params.partition || token.Partition || 'default',
      dc: params.dc,
      id: params.id,
    });

    // Remember which auth method started this login. The browser is about to
    // be redirected to the identity provider (see findCodeByURL below) and
    // when it comes back to /oidc/callback we need to know which method to
    // exchange the authorization code with.
    try {
      window.localStorage.setItem(
        'sabo:oidc:pending',
        JSON.stringify({
          authMethod: params.id,
          dc: params.dc,
          ns: params.ns || token.Namespace || 'default',
          partition: params.partition || token.Partition || 'default',
          apiPrefix: this.env.var('CONSUL_API_PREFIX') || '',
          redirectUri: `${this.env.var('CONSUL_BASE_UI_URL')}/oidc/callback`,
        })
      );
    } catch (e) {
      // ignore storage failures (for example in private browsing modes)
    }

    return res;
  }

  @dataSource('/:partition/:ns/:dc/oidc/authorize/:id/:code/:state')
  authorize(params, configuration = {}) {
    return this.store.authorize(this.getModelName(), params);
  }

  logout(id, code, state, dc, nspace, configuration = {}) {
    // TODO: Temporarily call this secret, as we alreayd do that with
    // self in the `store` look to see whether we should just call it id like
    // the rest
    const query = {
      id: id,
    };
    return this.store.logout(this.getModelName(), query);
  }

  close() {
    this.manager.close(OAUTH_PROVIDER_NAME);
  }

  findCodeByURL(src) {
    // Instead of opening a popup window (which browsers block because the
    // request that produced this URL is asynchronous and therefore no longer
    // counts as a user gesture) we navigate the whole window to the identity
    // provider. The provider then redirects back to
    // `${CONSUL_BASE_UI_URL}/oidc/callback`, which completes the login and
    // returns the user to the UI. See public/oidc/callback.
    set(this.provider, 'baseUrl', src);
    window.location.assign(src);
    // The page is navigating away. Return a promise that never settles so the
    // caller does not try to continue a flow that no longer exists.
    return new Promise(() => {});
  }
}

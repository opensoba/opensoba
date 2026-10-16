/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import Component from '@glimmer/component';
import { action } from '@ember/object';
import { tracked } from '@glimmer/tracking';

import chart from './chart.xstate';

export default class OidcSelect extends Component {
  @tracked items = [];
  @tracked error;
  @tracked provider;

  // Partitions are an OpenSoba Enterprise feature. We always talk to the
  // default partition; any value we send is stripped by the API client in
  // community builds.
  partition = 'default';

  constructor() {
    super(...arguments);
    this.chart = chart;
  }

  @action
  handleError(dispatch, e) {
    this.error = (e && (e.error || e.message)) || 'Failed to load SSO providers';
    dispatch('ERROR');
  }
}

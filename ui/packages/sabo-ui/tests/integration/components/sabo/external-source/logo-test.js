/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import { module, test } from 'qunit';
import { setupRenderingTest } from 'ember-qunit';
import { render } from '@ember/test-helpers';
import hbs from 'htmlbars-inline-precompile';

module('Integration | Component | sabo/external-source/logo', function (hooks) {
  setupRenderingTest(hooks);

  // The open source alternatives are not part of the Flight icon set, so each
  // one carries its own official mark: an inlined SVG, or a PNG for OpenWonton.
  test('it renders the mark of every open source alternative', async function (assert) {
    const expectedTag = {
      openbao: 'svg',
      opentofu: 'svg',
      openwonton: 'img',
      opendumpling: 'svg',
    };
    assert.expect(Object.keys(expectedTag).length * 2);

    for (const source of Object.keys(expectedTag)) {
      this.set('source', source);

      await render(hbs`<Sabo::ExternalSource::Logo @source={{this.source}} />`);

      const mark = this.element.querySelector('svg, img');
      assert.ok(mark, `${source} renders a mark`);
      assert.strictEqual(mark.tagName.toLowerCase(), expectedTag[source], `${source} mark type`);
    }
  });

  test('it inlines the official OpenDumpling mark without an accessible name', async function (assert) {
    this.set('source', 'opendumpling');

    await render(hbs`<Sabo::ExternalSource::Logo @source={{this.source}} />`);

    const svg = this.element.querySelector('svg');
    assert.ok(svg, 'renders an inline svg');
    assert.strictEqual(svg.getAttribute('viewBox'), '0 0 256 256');
    assert.strictEqual(svg.getAttribute('aria-hidden'), 'true');
    assert.notOk(svg.querySelector('title'), 'the name comes from the Registered via label');
    assert.ok(svg.querySelector('[fill="url(#odBody)"]'), 'the dumpling keeps its gradient fill');
  });

  test('it renders nothing for an unknown source', async function (assert) {
    this.set('source', 'not-supported');

    await render(hbs`<Sabo::ExternalSource::Logo @source={{this.source}} />`);

    assert.strictEqual(this.element.querySelector('svg, img'), null);
  });
});

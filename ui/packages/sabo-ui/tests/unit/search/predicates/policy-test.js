/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import { module, test } from 'qunit';

import ExactSearch from 'sabo-ui/utils/search/exact';
import predicates from 'sabo-ui/search/predicates/policy';

module('Unit | Search | Predicate | policy', function () {
  test('items are found by properties', function (assert) {
    const actual = new ExactSearch(
      [
        {
          Name: 'name-HIT',
          Description: 'description',
        },
        {
          Name: 'name',
          Description: 'desc-HIT-ription',
        },
      ],
      {
        finders: predicates,
      }
    ).search('hit');
    assert.equal(actual.length, 2);
  });
  test('items are not found', function (assert) {
    const actual = new ExactSearch(
      [
        {
          Name: 'name',
          Description: 'description',
        },
      ],
      {
        finders: predicates,
      }
    ).search('hit');
    assert.equal(actual.length, 0);
  });
});

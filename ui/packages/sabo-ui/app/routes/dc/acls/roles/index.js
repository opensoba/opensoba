/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import Route from 'sabo-ui/routing/route';
import { inject as service } from '@ember/service';

import WithBlockingActions from 'sabo-ui/mixins/with-blocking-actions';

export default class IndexRoute extends Route.extend(WithBlockingActions) {
  @service('repository/role') repo;
  queryParams = {
    sortBy: 'sort',
    searchproperty: {
      as: 'searchproperty',
      empty: [['Name', 'Description', 'Policy']],
    },
    search: {
      as: 'filter',
      replace: true,
    },
  };
}

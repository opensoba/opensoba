/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import StateService from 'sabo-ui/services/state';

import validate from 'sabo-ui/machines/validate.xstate';
import _boolean from 'sabo-ui/machines/boolean.xstate';

export default class ChartedStateService extends StateService {
  stateCharts = {
    validate: validate,
    boolean: _boolean,
  };
}

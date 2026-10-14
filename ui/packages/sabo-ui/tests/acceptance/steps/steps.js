/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import Inflector from 'ember-inflector';
import helpers from '@ember/test-helpers';
import $ from '-jquery';

import steps from 'sabo-ui/tests/steps';
import pages from 'sabo-ui/tests/pages';

import api from 'sabo-ui/tests/helpers/api';

export default function ({ assert, utils, library }) {
  return steps({
    assert,
    utils,
    library,
    pages,
    helpers,
    api,
    Inflector,
    $,
  });
}

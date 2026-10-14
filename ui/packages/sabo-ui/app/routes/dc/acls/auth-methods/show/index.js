/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import Route from 'sabo-ui/routing/route';
import to from 'sabo-ui/utils/routing/redirect-to';

export default class AuthMethodShowIndexRoute extends Route {
  redirect = to('auth-method');
}

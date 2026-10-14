/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import Serializer from './application';
import { PRIMARY_KEY, SLUG_KEY } from 'sabo-ui/models/role';
import WithPolicies from 'sabo-ui/mixins/policy/as-many';

export default class RoleSerializer extends Serializer.extend(WithPolicies) {
  primaryKey = PRIMARY_KEY;
  slugKey = SLUG_KEY;
}

/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import validations from 'sabo-ui/validations/role';
import builderFactory from 'sabo-ui/utils/form/builder';
const builder = builderFactory();
export default function (container, name = 'role', v = validations, form = builder) {
  return form(name, {}).setValidators(v).add(container.form('policy'));
}

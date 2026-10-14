/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import validations from 'sabo-ui/validations/kv';
import builderFactory from 'sabo-ui/utils/form/builder';
const builder = builderFactory();
export default function (container, name = '', v = validations, form = builder) {
  return form(name, {}).setValidators(v);
}

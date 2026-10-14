/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import Service from '@ember/service';
import btoa from 'sabo-ui/utils/btoa';
export default class BtoaService extends Service {
  execute() {
    return btoa(...arguments);
  }
}

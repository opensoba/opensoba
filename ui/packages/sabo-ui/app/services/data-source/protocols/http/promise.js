/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import Service from '@ember/service';
import { once } from 'sabo-ui/utils/dom/event-source';

export default class PromiseService extends Service {
  source(find, configuration) {
    return once(find, configuration);
  }
}

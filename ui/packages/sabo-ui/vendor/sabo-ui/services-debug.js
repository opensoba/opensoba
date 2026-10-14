/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

(services =>
  services({
    'route:application': {
      class: 'sabo-ui/routing/application-debug',
    },
    'service:intl': {
      class: 'sabo-ui/services/i18n-debug',
    },
  }))(
  (
    json,
    data = typeof document !== 'undefined' ? document.currentScript.dataset : module.exports
  ) => {
    data[`services`] = JSON.stringify(json);
  }
);

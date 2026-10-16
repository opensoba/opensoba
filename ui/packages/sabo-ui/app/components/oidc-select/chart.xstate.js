/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

// Providers are loaded as soon as the component is rendered (loading is the
// initial state) so there is no separate "enter a partition" step: partitions
// are an OpenSoba Enterprise feature and are not available here.
export default {
  id: 'oidc-select',
  initial: 'loading',
  states: {
    loading: {
      on: {
        SUCCESS: [
          {
            target: 'loaded',
          },
        ],
        ERROR: [
          {
            target: 'error',
          },
        ],
      },
    },
    loaded: {},
    error: {
      on: {
        RETRY: [
          {
            target: 'loading',
          },
        ],
      },
    },
  },
};

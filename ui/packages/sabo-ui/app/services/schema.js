/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import Service from '@ember/service';
import { schema as intentionPermissionSchema } from 'sabo-ui/models/intention-permission';
import { schema as intentionPermissionHttpSchema } from 'sabo-ui/models/intention-permission-http';
import { schema as intentionPermissionHttpHeaderSchema } from 'sabo-ui/models/intention-permission-http-header';

export default class SchemaService extends Service {
  'intention-permission' = intentionPermissionSchema;
  'intention-permission-http' = intentionPermissionHttpSchema;
  'intention-permission-http-header' = intentionPermissionHttpHeaderSchema;
}

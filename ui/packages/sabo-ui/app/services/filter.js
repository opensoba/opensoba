/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import Service from '@ember/service';
import { andOr } from 'sabo-ui/utils/filter';

import service from 'sabo-ui/filter/predicates/service';
import serviceInstance from 'sabo-ui/filter/predicates/service-instance';
import healthCheck from 'sabo-ui/filter/predicates/health-check';
import node from 'sabo-ui/filter/predicates/node';
import kv from 'sabo-ui/filter/predicates/kv';
import intention from 'sabo-ui/filter/predicates/intention';
import token from 'sabo-ui/filter/predicates/token';
import policy from 'sabo-ui/filter/predicates/policy';
import authMethod from 'sabo-ui/filter/predicates/auth-method';
import peer from 'sabo-ui/filter/predicates/peer';

const predicates = {
  service: andOr(service),
  ['service-instance']: andOr(serviceInstance),
  ['health-check']: andOr(healthCheck),
  ['auth-method']: andOr(authMethod),
  node: andOr(node),
  kv: andOr(kv),
  intention: andOr(intention),
  token: andOr(token),
  policy: andOr(policy),
  peer: andOr(peer),
};

export default class FilterService extends Service {
  predicate(type) {
    return predicates[type];
  }
}

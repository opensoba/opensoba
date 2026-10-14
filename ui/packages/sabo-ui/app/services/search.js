/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import Service from '@ember/service';

import ExactSearch from 'sabo-ui/utils/search/exact';

import intention from 'sabo-ui/search/predicates/intention';
import upstreamInstance from 'sabo-ui/search/predicates/upstream-instance';
import serviceInstance from 'sabo-ui/search/predicates/service-instance';
import healthCheck from 'sabo-ui/search/predicates/health-check';
import acl from 'sabo-ui/search/predicates/acl';
import service from 'sabo-ui/search/predicates/service';
import node from 'sabo-ui/search/predicates/node';
import kv from 'sabo-ui/search/predicates/kv';
import token from 'sabo-ui/search/predicates/token';
import role from 'sabo-ui/search/predicates/role';
import policy from 'sabo-ui/search/predicates/policy';
import authMethod from 'sabo-ui/search/predicates/auth-method';
import nspace from 'sabo-ui/search/predicates/nspace';
import peer from 'sabo-ui/search/predicates/peer';

const predicates = {
  intention: intention,
  service: service,
  ['service-instance']: serviceInstance,
  ['upstream-instance']: upstreamInstance,
  ['health-check']: healthCheck,
  ['auth-method']: authMethod,
  node: node,
  kv: kv,
  acl: acl,
  token: token,
  role: role,
  policy: policy,
  nspace: nspace,
  peer: peer,
};

export default class SearchService extends Service {
  searchables = {
    exact: ExactSearch,
    // regex: RegExpSearch,
    // fuzzy: FuzzySearch,
  };
  predicate(name) {
    return predicates[name];
  }
}

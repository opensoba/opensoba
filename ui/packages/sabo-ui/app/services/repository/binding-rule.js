/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import RepositoryService from 'sabo-ui/services/repository';
import { PRIMARY_KEY, SLUG_KEY } from 'sabo-ui/models/binding-rule';
import dataSource from 'sabo-ui/decorators/data-source';

const MODEL_NAME = 'binding-rule';

export default class BindingRuleService extends RepositoryService {
  getModelName() {
    return MODEL_NAME;
  }

  getPrimaryKey() {
    return PRIMARY_KEY;
  }

  getSlugKey() {
    return SLUG_KEY;
  }

  @dataSource('/:partition/:ns/:dc/binding-rules/for-auth-method/:authmethod')
  async findAllByAuthMethod() {
    return super.findAll(...arguments);
  }
}

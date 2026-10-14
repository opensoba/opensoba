/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import RepositoryService from 'sabo-ui/services/repository';
import dataSource from 'sabo-ui/decorators/data-source';

const modelName = 'node';
export default class NodeService extends RepositoryService {
  getModelName() {
    return modelName;
  }

  @dataSource('/:partition/:ns/:dc/nodes')
  async findAllByDatacenter() {
    return super.findAllByDatacenter(...arguments);
  }

  @dataSource('/:partition/:ns/:dc/node/:id/:peer')
  async findBySlug() {
    return super.findBySlug(...arguments);
  }

  @dataSource('/:partition/:ns/:dc/leader')
  findLeader(params, configuration = {}) {
    if (typeof configuration.refresh !== 'undefined') {
      params.uri = configuration.uri;
    }
    return this.store.queryLeader(this.getModelName(), params);
  }
}

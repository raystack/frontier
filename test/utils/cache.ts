import { createCacheId } from '../../src/utils/cache';
import Code from 'code';

import Lab from '@hapi/lab';
import Sinon from 'sinon';
import { lab } from '../setup';

exports.lab = Lab.script();

lab.experiment('Cache::createCacheId', () => {
  lab.test('should return empty string', () => {
    const cacheId = createCacheId();
    Code.expect(cacheId).to.equal('');
  });

  lab.test('should return query string of object', () => {
    const obj = {
      entity: 'entity1',
      landscape: 'landscape1'
    };
    const cacheId = createCacheId(obj);
    Code.expect(cacheId).to.equal('entity=entity1&landscape=landscape1');
  });

  lab.test('should return cacheId of nested object', () => {
    const obj = {
      entity: 'entity1',
      filters: {
        filter1: 'filter1',
        filter2: 'filter2'
      }
    };
    const cacheId = createCacheId(obj);
    Code.expect(cacheId).to.equal(
      'entity=entity1&filters.filter1=filter1&filters.filter2=filter2'
    );
  });

  lab.test('should return cacheId of nested array', () => {
    const obj = {
      entity: 'entity1',
      filters: [{ id: 'filter1' }, { id: 'filter2' }]
    };
    const cacheId = createCacheId(obj);
    Code.expect(cacheId).to.equal(
      'entity=entity1&filters[0].id=filter1&filters[1].id=filter2'
    );
  });

  lab.test('should add userId to cacheId', () => {
    const obj = {
      entity: 'entity1',
      filters: [{ id: 'filter1' }, { id: 'filter2' }]
    };
    const cacheId = createCacheId(obj, 'user1');
    Code.expect(cacheId).to.equal(
      'user1::entity=entity1&filters[0].id=filter1&filters[1].id=filter2'
    );
  });

  lab.test('should add name to cacheId', () => {
    const obj = {
      entity: 'entity1',
      filters: [{ id: 'filter1' }, { id: 'filter2' }]
    };
    const cacheId = createCacheId(obj, 'user1', 'group');
    Code.expect(cacheId).to.equal(
      'user1::group::entity=entity1&filters[0].id=filter1&filters[1].id=filter2'
    );
  });
});

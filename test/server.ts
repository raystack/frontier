/* eslint-disable no-underscore-dangle */

import Code from 'code';
import Lab from '@hapi/lab';
import Composer from '../config/composer';

const lab = Lab.script();
exports.lab = lab;

lab.experiment('Server', () => {
  lab.test('starts the hapi server', async () => {
    const server = await Composer();
    await server.start();
    Code.expect(server._core.phase).to.equal('started');
    await server.stop();
    Code.expect(server._core.phase).to.equal('stopped');
  });
});

import Laabr from 'laabr';
import * as R from 'ramda';

const config = {
  plugin: Laabr,
  options: {
    colored: true,
    hapiPino: {
      ignorePaths: ['/ping']
    },
    tokens: {
      path: (data: any) => {
        const { url } = data.req;
        return `${url.pathname}${url.search}${url.hash}`;
      }
    },
    preformatter: (data: any) => {
      const statusCode = R.pathOr(null, ['res', 'statusCode'], data);
      if (statusCode && statusCode < 400) {
        // eslint-disable-next-line no-param-reassign
        data.payload = '';
        return data;
      }
      return data;
    },
    formats: {
      onPostStart: ':time[iso] [:level] :message at: :host[uri]',
      onPostStop: ':time[iso] [:level] :message at: :host[uri]',
      log: ':time[iso] [:level] :tags :message',
      request: ':time[iso] [:level] :message',
      response:
        ':time[iso] [:level] :method :path :status :payload (:responseTime ms)',
      uncaught:
        ':time[iso] [:level] :method :path :payload :error[source] :error[stack]',
      'request-error':
        ':time[iso] [:level] :method :path :payload :error[message] :error[stack]'
    }
  }
};

export default config;

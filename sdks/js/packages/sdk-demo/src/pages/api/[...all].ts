import type { NextApiRequest, NextApiResponse } from 'next';
import httpProxyMiddleware from 'next-http-proxy-middleware';

export const config = {
  api: {
    // Enable `externalResolver` option in Next.js
    externalResolver: true
  }
};

// eslint-disable-next-line import/no-anonymous-default-export
export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse
) {
  const baseUrl = process.env.FRONTIER_ENDPOINT || 'http://frontier:8080';

  await httpProxyMiddleware(req, res, {
    // You can use the `http-proxy` option
    target: baseUrl,
    pathRewrite: [
      {
        patternStr: '^/api',
        replaceStr: ''
      }
    ]
  });
}

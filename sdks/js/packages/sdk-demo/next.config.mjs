/** @type {import('next').NextConfig} */

const baseUrl = process.env.FRONTIER_ENDPOINT || 'http://frontier:8080'

const nextConfig = {
  rewrites: async () => {
    return [
      {
        source: '/api/:path*',
        destination: `${baseUrl}/:path*`
      }
    ]
  }
};

export default nextConfig;

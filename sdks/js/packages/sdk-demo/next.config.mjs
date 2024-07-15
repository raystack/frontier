/** @type {import('next').NextConfig} */

const baseUrl = process.env.FRONTIER_ENDPOINT || 'http://localhost:8000'
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

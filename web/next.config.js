/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  reactStrictMode: false,
  images: {
    unoptimized: true,
  },
  typescript: {
    tsconfigPath: './tsconfig.json',
  },
  transpilePackages: ["@fluentui/react-components"],
  ...(process.env.NODE_ENV === 'development' ? {
    async rewrites() {
      return {
        beforeFiles: [
          {
            source: '/api/:path*',
            destination: 'http://localhost:2703/api/:path*',
          },
        ],
      };
    }
  } : {})
};

export default nextConfig;

const ENV_CONFIG = {
    API_BASE_URL: process.env.NEXT_PUBLIC_API_URL,
    DOMAIN: process.env.NEXT_PUBLIC_DOMAIN,
    // These are exposed via next.config.ts from GITHUB_CLIENT_ID and GOOGLE_CLIENT_ID
    GITHUB_CLIENT_ID: process.env.NEXT_PUBLIC_GITHUB_CLIENT_ID,
    GOOGLE_CLIENT_ID: process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID,
};

export default ENV_CONFIG;


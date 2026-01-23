import { ApiClient } from '@raystack/frontier/api-client'
import config from '@/config/frontier';

const client = new ApiClient({
    baseUrl: config.endpoint,
    baseApiParams: {
        credentials: 'include'
    }
});

export default client
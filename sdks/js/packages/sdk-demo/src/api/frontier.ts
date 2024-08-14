import { ApiClient } from '@raystack/frontier/apiClient'
import config from '@/config/frontier';

const client = new ApiClient({
    baseUrl: config.endpoint,
    baseApiParams: {
        credentials: 'include'
    }
});

export default client
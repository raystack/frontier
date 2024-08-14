import { ApiClient } from '../../../core/ApiClient/dist'
import config from '@/config/frontier';

const client = new ApiClient({
    baseUrl: config.endpoint,
    baseApiParams: {
        credentials: 'include'
    }
});

export default client
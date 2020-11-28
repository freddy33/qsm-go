import axios from 'axios';
import localStorage from './localStorage';
import { LOCAL_STORAGE_KEY } from '../constant';

const request = (options) => {
  const envId = localStorage.getItem(LOCAL_STORAGE_KEY.SELECTED_ENVIRONMENT) || 1;

  return axios({
    headers: {
      QsmEnvId: envId,
    },
    ...options,
  });
};

const get = (url, params) => {
  return request({
    method: 'get',
    url,
    params,
  });
};

const post = (url, data) => {
  return request({
    method: 'post',
    url,
    data,
  });
};

const put = (url, data) => {
  return request({
    method: 'put',
    url,
    data,
  });
};

// variable name "delete" is reserved, so need to use _delete here
const _delete = (url, data) => {
  return request({
    method: 'delete',
    url,
    data,
  });
};

export default {
  request,
  get,
  post,
  put,
  delete: _delete,
};

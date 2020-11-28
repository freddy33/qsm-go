import { LOCAL_STORAGE_KEY } from '../constant';

const constructKey = (key) => `qsm:${key}`;

const setItem = (key, value) => {
  localStorage.setItem(constructKey(key), value);
};

const getItem = (key) => {
  return localStorage.getItem(constructKey(key));
};

const getCurrentEnv = () => {
  return getItem(LOCAL_STORAGE_KEY.SELECTED_ENVIRONMENT);
};

export default {
  setItem,
  getItem,
  getCurrentEnv,
};

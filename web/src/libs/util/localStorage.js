const constructKey = (key) => `qsm:${key}`;

const setItem = (key, value) => {
  localStorage.setItem(constructKey(key), value);
};

const getItem = (key) => {
  return localStorage.getItem(constructKey(key));
};

export default {
  setItem,
  getItem,
};

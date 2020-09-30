import axios from 'axios';

import m3point from '../grpc/m3point_pb';

const fetchPointPackDataMsg = async () => {
  const resp = await axios.get('./mock/point-data', {
    responseType: 'arraybuffer',
  });

  const pointPackData = m3point.PointPackDataMsg.deserializeBinary(resp.data);

  return pointPackData;
};

export default {
  fetchPointPackDataMsg,
};

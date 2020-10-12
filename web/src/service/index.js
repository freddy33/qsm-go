import _ from 'lodash';
import axios from 'axios';

import m3point from '../grpc/m3point_pb';

const REACT_APP_BACKEND_URL = process.env.REACT_APP_BACKEND_URL;

const fetchPointPackDataMsgGrpc = async () => {
  const resp = await axios.get(`${REACT_APP_BACKEND_URL}/point-data`, {
    responseType: 'arraybuffer',
  });

  const pointPackDataMsg = m3point.PointPackDataMsg.deserializeBinary(resp.data);

  const connections = {};
  const trios = {};
  pointPackDataMsg.getAllConnectionsList().forEach((conn) => {
    connections[conn.getConnId()] = {
      connId: conn.getConnId(),
      ds: conn.getDs(),
      vector: {
        x: conn.getVector().getX(),
        y: conn.getVector().getY(),
        z: conn.getVector().getZ(),
      },
    };
  });

  pointPackDataMsg.getAllTriosList().forEach((trio) => {
    trios[trio.getTrioId()] = {
      trioId: trio.getTrioId(),
      connIds: trio.getConnIdsList(),
    };
  });

  return { connections, trios };
};

const fetchPointPackDataMsg = async () => {
  const resp = await axios({
    method: 'get',
    url: `${REACT_APP_BACKEND_URL}/point-data`,
    data: null,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  const connections = {};
  const trios = {};
  const pointPackDataMsg = _.get(resp, 'data', {});

  const allConnections = _.get(pointPackDataMsg, 'all_connections', []);
  allConnections.forEach((conn) => {
    connections[_.get(conn, 'conn_id', 0)] = {
      connId: _.get(conn, 'conn_id', 0),
      ds: _.get(conn, 'ds', 0),
      vector: {
        x: _.get(conn, 'vector.x', 0),
        y: _.get(conn, 'vector.y', 0),
        z: _.get(conn, 'vector.z', 0),
      },
    };
  });

  const allTrios = _.get(pointPackDataMsg, 'all_trios', []);
  allTrios.forEach((trio) => {
    trios[_.get(trio, 'trio_id', 0)] = {
      trioId: _.get(trio, 'trio_id', 0),
      connIds: _.get(trio, 'conn_ids', []),
    };
  });

  return { connections, trios };
};

const initEnv = async () => {
  const resp = await axios({
    method: 'post',
    url: `${REACT_APP_BACKEND_URL}/init-env`,
    data: null,
    headers: {
      'Content-Type': 'application/json',
    },
  });
};

const createPathContext = async (growthType, growthIndex, growthOffset) => {
  const resp = await axios({
    method: 'post',
    url: `${REACT_APP_BACKEND_URL}/path-context`,
    data: {
      growth_type: growthType,
      growth_index: growthIndex,
      growth_offset: growthOffset,
    },
    headers: {
      'Content-Type': 'application/json',
    },
  });

  return resp.data;
};

const updateMaxDist = async (pathContextId, dist) => {
  const resp = await axios({
    method: 'put',
    url: `${REACT_APP_BACKEND_URL}/max-dist`,
    data: {
      path_ctx_id: pathContextId,
      dist,
    },
    headers: {
      'Content-Type': 'application/json',
    },
  });

  return resp.data;
};

const getPathNodes = async (pathContextId, toDist) => {
  const resp = await axios({
    method: 'get',
    url: `${REACT_APP_BACKEND_URL}/path-nodes`,
    params: {
      path_ctx_id: pathContextId,
      dist: 0,
      to_dist: toDist,
    },
  });

  return resp.data;
};

export default {
  fetchPointPackDataMsg,
  fetchPointPackDataMsgGrpc,
  initEnv,
  createPathContext,
  updateMaxDist,
  getPathNodes,
};

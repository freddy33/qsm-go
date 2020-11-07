import _ from 'lodash';
import axios from 'axios';

const REACT_APP_BACKEND_URL = process.env.REACT_APP_BACKEND_URL;

const getPointPackDataMsg = async () => {
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

const updateMaxDist = async (pathContextId, dist) => {
  const resp = await axios({
    method: 'put',
    url: `${REACT_APP_BACKEND_URL}/max-dist`,
    data: {
      path_ctx_id: pathContextId,
      dist,
    },
  });

  const maxDist = _.get(resp, 'data.max_dist');
  if (!maxDist) {
    alert(resp.data);
  }
};

const getPathNodes = async (pathContextId, fromDist, toDist) => {
  const resp = await axios({
    method: 'get',
    url: `${REACT_APP_BACKEND_URL}/path-nodes`,
    params: {
      path_ctx_id: pathContextId,
      dist: fromDist,
      to_dist: toDist,
    },
  });

  return resp.data;
};

const getPathContext = async (pathContextId) => {
  const resp = await axios({
    method: 'get',
    url: `${REACT_APP_BACKEND_URL}/path-context`,
    params: {
      path_ctx_id: pathContextId,
    },
  });

  const pathContext = _.get(resp, 'data', {});
  const { path_ctx_id, growth_type, growth_index, growth_offset, max_dist } = pathContext;

  return {
    pathContextId: path_ctx_id,
    growthType: growth_type,
    growthIndex: growth_index,
    growthOffset: growth_offset,
    maxDist: max_dist,
  };
};

const getPathContextIds = async () => {
  const resp = await axios({
    method: 'get',
    url: `${REACT_APP_BACKEND_URL}/path-context`,
    params: {
      path_ctx_id: -1,
    },
  });

  const pathContexts = _.get(resp, 'data.path_contexts', []);

  const pathContextIds = pathContexts.map((pathContext) => {
    return pathContext.path_ctx_id;
  });

  const sorted = _.sortBy(pathContextIds);
  return sorted;
};

export default {
  getPointPackDataMsg,
  updateMaxDist,
  getPathNodes,
  getPathContext,
  getPathContextIds,
};

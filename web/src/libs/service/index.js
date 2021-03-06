import _ from 'lodash';
import httpRequest from '../util/httpRequest';
import { HTTP_STATUS } from '../constant';

const REACT_APP_BACKEND_URL = process.env.REACT_APP_BACKEND_URL;

const getPointPackDataMsg = async () => {
  const resp = await httpRequest.get(`${REACT_APP_BACKEND_URL}/point-data`);

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

const createPathContext = async (growthType, growthIndex, growthOffset) => {
  const resp = await httpRequest.post(`${REACT_APP_BACKEND_URL}/path-context`, {
    growth_type: growthType,
    growth_index: growthIndex,
    growth_offset: growthOffset,
  });

  if (!_.get(resp, 'data.path_ctx_id')) {
    alert(resp.data);
  }

  const data = _.get(resp, 'data');
  return data;
};

const updateMaxDist = async (pathContextId, dist) => {
  const resp = await httpRequest.put(`${REACT_APP_BACKEND_URL}/max-dist`, {
    path_ctx_id: pathContextId,
    dist,
  });

  const maxDist = _.get(resp, 'data.max_dist');
  if (!maxDist) {
    alert(resp.data);
  }
};

const getPathNodes = async (pathContextId, fromDist, toDist) => {
  const resp = await httpRequest.get(`${REACT_APP_BACKEND_URL}/path-nodes`, {
    path_ctx_id: pathContextId,
    dist: fromDist,
    to_dist: toDist,
  });

  return resp.data;
};

const getPathContext = async (pathContextId) => {
  const resp = await httpRequest.get(`${REACT_APP_BACKEND_URL}/path-context`, {
    path_ctx_id: pathContextId,
  });

  const pathContext = _.get(resp, 'data', {});
  const {
    path_ctx_id,
    growth_type,
    growth_index,
    growth_offset,
    max_dist,
  } = pathContext;

  return {
    pathContextId: path_ctx_id,
    growthType: growth_type,
    growthIndex: growth_index,
    growthOffset: growth_offset,
    maxDist: max_dist,
  };
};

const getPathContexts = async () => {
  const resp = await httpRequest.get(`${REACT_APP_BACKEND_URL}/path-context`, {
    path_ctx_id: -1,
  });

  const pathContexts = _.get(resp, 'data.path_contexts', []);
  return pathContexts;
};

const getPathContextIds = async () => {
  const pathContexts = await getPathContexts();

  const pathContextIds = pathContexts.map((pathContext) => {
    return pathContext.path_ctx_id;
  });

  const sorted = _.sortBy(pathContextIds);
  return sorted;
};

const getEnvironments = async () => {
  const resp = await httpRequest.get(`${REACT_APP_BACKEND_URL}/list-env`);

  return _.get(resp, 'data.envs', []);
};

const createEnvironment = async (envId) => {
  const resp = await httpRequest.request({
    method: 'post',
    url: `${REACT_APP_BACKEND_URL}/init-env`,
    headers: {
      QsmEnvId: envId,
    },
  });

  if (_.get(resp, 'status') !== 201) {
    alert(resp.data);
  }
};

const deleteEnvironment = async (envId) => {
  const resp = await httpRequest.request({
    method: 'delete',
    url: `${REACT_APP_BACKEND_URL}/drop-env`,
    headers: {
      QsmEnvId: envId,
    },
  });

  if (_.get(resp, 'status') !== HTTP_STATUS.OK) {
    alert(resp.data);
  }
};

const getSpaces = async () => {
  const resp = await httpRequest.get(`${REACT_APP_BACKEND_URL}/space`);

  return _.get(resp, 'data.spaces', []);
};

const createSpace = async (spaceName, activeThreshold, maxTime, maxCoord) => {
  try {
    const resp = await httpRequest.post(`${REACT_APP_BACKEND_URL}/space`, {
      space_name: spaceName,
      active_threshold: _.parseInt(activeThreshold),
      max_time: _.parseInt(maxTime),
      max_coord: _.parseInt(maxCoord),
    });

    if (_.get(resp, 'status') !== HTTP_STATUS.OK) {
      alert(resp.data);
    }

    return _.get(resp, 'data');
  } catch (e) {
    alert(_.get(e, 'response.data'));
  }
};

const deleteSpace = async (spaceId, spaceName) => {
  try {
    const resp = await httpRequest.delete(`${REACT_APP_BACKEND_URL}/space`, {
      space_id: spaceId,
      space_name: spaceName,
    });

    if (_.get(resp, 'status') !== HTTP_STATUS.OK) {
      alert(resp.data);
    }

    return _.get(resp, 'data');
  } catch (e) {
    alert(_.get(e, 'response.data'));
  }
};

const getEvents = async (spaceId) => {
  try {
    const resp = await httpRequest.get(`${REACT_APP_BACKEND_URL}/event`, {
      space_id: spaceId,
      at_time: -1,
    });

    return _.get(resp, 'data.events', []);
  } catch (e) {
    alert(_.get(e, 'response.data'));
  }
};

const createEvent = async (
  spaceId,
  growthType,
  growthIndex,
  growthOffset,
  creationTime,
  centerX,
  centerY,
  centerZ,
  color,
) => {
  try {
    const resp = await httpRequest.post(`${REACT_APP_BACKEND_URL}/event`, {
      space_id: _.parseInt(spaceId),
      growth_type: _.parseInt(growthType),
      growth_index: _.parseInt(growthIndex),
      growth_offset: _.parseInt(growthOffset),
      creationTime: _.parseInt(creationTime),
      center: {
        x: _.parseInt(centerX),
        y: _.parseInt(centerY),
        z: _.parseInt(centerZ),
      },
      color: _.parseInt(color),
    });

    return _.get(resp, 'data');
  } catch (e) {
    alert(_.get(e, 'response.data'));
  }
};

const getSpaceTime = async (
  spaceId,
  currentTime,
  minNbEventsFilter,
  colorMaskFilter,
) => {
  try {
    const resp = await httpRequest.get(`${REACT_APP_BACKEND_URL}/space-time`, {
      space_id: spaceId,
      current_time: currentTime,
      min_nb_events_filter: minNbEventsFilter,
      color_mask_filter: colorMaskFilter,
    });

    return _.get(resp, 'data');
  } catch (e) {
    alert(_.get(e, 'response.data'));
  }
};

export default {
  getPointPackDataMsg,
  createPathContext,
  updateMaxDist,
  getPathNodes,
  getPathContext,
  getPathContexts,
  getPathContextIds,
  getEnvironments,
  createEnvironment,
  deleteEnvironment,
  getSpaces,
  createSpace,
  deleteSpace,
  getEvents,
  createEvent,
  getSpaceTime,
};

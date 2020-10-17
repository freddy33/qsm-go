import React, { useEffect, useState, useRef } from 'react';
import * as THREE from 'three';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls';
import _ from 'lodash';
import Select from 'react-select';

import Service from './service';
import Renderer from './renderer';

const growthTypeOptions = [1, 2, 3, 4, 8].map((v) => ({ value: v, label: v }));
const growthIndexOptions = [...Array(12).keys()].map((v) => ({ value: v, label: v }));
const growthOffsetOptions = [...Array(12).keys()].map((v) => ({ value: v, label: v }));

const App = () => {
  const mount = useRef(null);

  const [scene, setScene] = useState();
  const [camera, setCamera] = useState();
  const [renderer, setRenderer] = useState();
  const [pointPackDataMsg, setPointPackDataMsg] = useState();
  const [growthTypeOption, setGrowthTypeOption] = useState(_.last(growthTypeOptions));
  const [growthIndexOption, setGrowthIndexOption] = useState(_.first(growthIndexOptions));
  const [growthOffsetOption, setGrowthOffsetOption] = useState(_.first(growthOffsetOptions));
  const [currentPathContextId, setCurrentPathContextId] = useState();
  const [maxDist, setMaxDist] = useState(0);
  const [drawingRoots, setDrawingRoots] = useState([]);
  const [getPathNodesRequest, setGetPathNodeRequest] = useState({ fromDist: 0, toDist: 0 });

  const fetchPointPackDataMsg = () => {
    Service.fetchPointPackDataMsg().then((pointPackDataMsg) => {
      setPointPackDataMsg(pointPackDataMsg);
    });
  };

  const createPathContext = async () => {
    const resp = await Service.createPathContext(
      growthTypeOption.value,
      growthIndexOption.value,
      growthOffsetOption.value
    );

    const point = _.get(resp, 'root_path_node.point');
    if (!point) return;
    Renderer.addPoint(scene, point);

    const pathContextId = _.get(resp, 'path_ctx_id');
    setCurrentPathContextId(pathContextId);

    const maxDist = _.get(resp, 'max_dist', 0);
    setMaxDist(maxDist);
  };

  const updateMaxDist = async () => {
    const resp = await Service.updateMaxDist(currentPathContextId, parseInt(maxDist));

    const dist = _.get(resp, 'max_dist');
    if (!dist) {
      alert(resp);
    }

    setMaxDist(dist);
  };

  const getPathNodes = async () => {
    const fromDist = _.get(getPathNodesRequest, 'fromDist', 0);
    const toDist = _.get(getPathNodesRequest, 'toDist', 0);
    const resp = await Service.getPathNodes(currentPathContextId, fromDist, toDist);
    const pathNodes = _.get(resp, 'path_nodes', []);

    if (!pathNodes) {
      alert(resp);
    }

    const sortByDist = _.sortBy(pathNodes, ['d']);

    const nodeMap = {};
    sortByDist.forEach((node) => {
      const linkedPathNodeIds = _.get(node, 'linked_path_node_ids', []);

      linkedPathNodeIds.forEach((precedingNodeId) => {
        const precedingNode = nodeMap[precedingNodeId];

        if (!precedingNode) {
          return;
        }

        const childNodes = _.get(precedingNode, 'childNodes', []);
        childNodes.push(node);
        precedingNode.childNodes = childNodes;
      });

      const pathNodeId = node.path_node_id;
      nodeMap[pathNodeId] = node;
    });

    const roots = _.filter(nodeMap, { d: fromDist });

    setDrawingRoots(roots);
  };

  // componentDidMount, will load once only when page start
  useEffect(() => {
    fetchPointPackDataMsg();

    const { clientWidth: width, clientHeight: height } = mount.current;

    const { scene, camera, renderer } = Renderer.init(width, height);
    setScene(scene);
    setCamera(camera);
    setRenderer(renderer);

    mount.current.appendChild(renderer.domElement);

    const handleResize = () => {
      const { clientWidth: width, clientHeight: height } = mount.current;
      renderer.setSize(width, height);
      camera.aspect = width / height;
      camera.updateProjectionMatrix();
      renderer.render(scene, camera);
    };

    const controls = new OrbitControls(camera, renderer.domElement);

    window.addEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    if (!renderer) return;

    const { clientWidth: width, clientHeight: height } = mount.current;

    const { scene, camera, renderer: newRenderer } = Renderer.init(width, height);
    setScene(scene);
    setCamera(camera);
    setRenderer(newRenderer);

    mount.current.replaceChild(newRenderer.domElement, renderer.domElement);
    const controls = new OrbitControls(camera, newRenderer.domElement);

    Renderer.drawRoots(scene, drawingRoots);
  }, [drawingRoots]);

  // called for every button clicks to update how the UI should render
  useEffect(() => {
    if (!(scene && camera && renderer)) return;

    const cameraPivot = Renderer.addCameraPivot(scene, camera);
    const animate = () => {
      // cameraPivot.rotateOnAxis(new THREE.Vector3(0, 1, 0), 0.01);

      requestAnimationFrame(animate);
      renderer.render(scene, camera);
    };

    animate();
  }, [scene, camera, renderer]);

  return (
    <div className="main">
      <div className="panel">
        <div>
          <span>Growth Type</span>
          <Select defaultValue={growthTypeOption} onChange={setGrowthTypeOption} options={growthTypeOptions} />
          <span>Growth Index</span>
          <Select defaultValue={growthIndexOption} onChange={setGrowthIndexOption} options={growthIndexOptions} />
          <span>Growth Offset</span>
          <Select defaultValue={growthOffsetOption} onChange={setGrowthOffsetOption} options={growthOffsetOptions} />
        </div>
        <div>
          <button onClick={() => createPathContext()}>Create Path Context</button>
        </div>

        <hr />
        <div>
          <span>Max Dist: </span>
          <input
            type="number"
            value={maxDist}
            onChange={(evt) => {
              setMaxDist(evt.target.value);
            }}
          />
        </div>
        <div>
          <button disabled={!currentPathContextId} onClick={() => updateMaxDist()}>
            Update Max Dist
          </button>
        </div>
        <hr />
        <div>
          <div>
            <span>From Dist: </span>
            <input
              type="number"
              value={_.get(getPathNodesRequest, 'fromDist', 0)}
              onChange={(evt) => {
                setGetPathNodeRequest({ ...getPathNodesRequest, fromDist: parseInt(evt.target.value) });
              }}
            />
          </div>
          <div>
            <span>To Dist: </span>
            <input
              type="number"
              value={_.get(getPathNodesRequest, 'toDist', 0)}
              onChange={(evt) => {
                setGetPathNodeRequest({ ...getPathNodesRequest, toDist: parseInt(evt.target.value) });
              }}
            />
          </div>
          <div>
            <button disabled={!currentPathContextId} onClick={() => getPathNodes()}>
              Get Path Nodes (Redraw)
            </button>
          </div>
        </div>
      </div>
      <div className="vis" ref={mount} />
    </div>
  );
};

export default App;

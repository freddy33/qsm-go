import React, { useEffect, useState, useRef } from 'react';
import * as THREE from 'three';
import _ from 'lodash';
import Select from 'react-select';

import Service from './service';
import Renderer from './renderer';

const growthTypeOptions = [1, 2, 3, 4, 8].map((v) => ({ value: v, label: v }));
const growthIndexOptions = [...Array(12).keys()].map((v) => ({ value: v, label: v }));
const growthOffsetOptions = [...Array(12).keys()].map((v) => ({ value: v, label: v }));

const App = () => {
  const mount = useRef(null);

  const [rotating, setRotating] = useState(true);
  const [scene, setScene] = useState();
  const [camera, setCamera] = useState();
  const [renderer, setRenderer] = useState();
  const [pointPackDataMsg, setPointPackDataMsg] = useState();
  const [dataInput, setDataInput] = useState('');
  const [growthTypeOption, setGrowthTypeOption] = useState(_.last(growthTypeOptions));
  const [growthIndexOption, setGrowthIndexOption] = useState(_.first(growthIndexOptions));
  const [growthOffsetOption, setGrowthOffsetOption] = useState(_.first(growthOffsetOptions));
  const [currentPathContextId, setCurrentPathContextId] = useState();
  const [maxDist, setMaxDist] = useState(0);

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
  };

  const updateMaxDist = async () => {
    const resp = await Service.updateMaxDist(currentPathContextId, maxDist + 1);

    const dist = _.get(resp, 'max_dist');
    setMaxDist(maxDist);
  };

  const buildTree = (root, nodeMap) => {
    const childNodes = _.get(root, 'childNodes', []);
    if (!childNodes.length) {
      return root;
    }

    root.childNodes = childNodes.map((child) => {
      return buildTree(child, nodeMap);
    });
  };

  const getPathNodes = async () => {
    // const resp = await Service.getPathNodes(currentPathContextId, maxDist);
    const resp = await Service.getPathNodes(105, 2);
    const pathNodes = _.get(resp, 'path_nodes', []);

    // build tree
    // const groupByDist = _.groupBy(pathNodes, 'd')
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

    const root = nodeMap[sortByDist[0].path_node_id];
    const nodeTree = buildTree(root, nodeMap);

    debugger;
    console.log(nodeTree);
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

    window.addEventListener('resize', handleResize);

    setDataInput(JSON.stringify(Renderer.mockPoints));
  }, []);

  // useEffect(() => {
  //   Renderer.draw(scene, Renderer.mockPoints, pointPackDataMsg);
  // }, [pointPackDataMsg]);

  // called for every button clicks to update how the UI should render
  useEffect(() => {
    if (!(scene && camera && renderer)) return;

    const cameraPivot = Renderer.addCameraPivot(scene, camera);
    const animate = () => {
      if (rotating) {
        cameraPivot.rotateOnAxis(new THREE.Vector3(0, 1, 0), 0.01);
      }

      requestAnimationFrame(animate);
      renderer.render(scene, camera);
    };

    animate();
  }, [rotating, scene, camera, renderer]);

  return (
    <div className="main">
      <div className="panel">
        {/* <div>
          <textarea
            onChange={(evt) => {
              setDataInput(_.get(evt, 'target.value', ''));
            }}
            rows="15"
            value={dataInput}
          />
        </div>
        <div>
          <button
            onClick={() => {
              const data = JSON.parse(dataInput);

              const { clientWidth: width, clientHeight: height } = mount.current;

              const { scene, camera, renderer: newRenderer } = Renderer.init(width, height);
              setScene(scene);
              setCamera(camera);
              setRenderer(newRenderer);

              mount.current.replaceChild(newRenderer.domElement, renderer.domElement);
              Renderer.draw(scene, data, pointPackDataMsg);
            }}
          >
            Load data
          </button>
        </div>
        <hr /> */}
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

        <div>
          <button onClick={() => getPathNodes()}>Get Path Context</button>
        </div>
        <hr />
        <div>
          <button onClick={() => updateMaxDist()}>Update Max Dist</button>
        </div>
        <hr />
        <div>
          <button onClick={() => Service.initEnv()}>Init Env</button>
        </div>
        <div>
          <button onClick={() => fetchPointPackDataMsg()}>Fetch PointPackDataMsg</button>
        </div>
        {/* <div>
          <button onClick={() => setRotating(!rotating)}>Rotate</button>
        </div> */}
      </div>
      <div className="vis" ref={mount} />
    </div>
  );
};

export default App;

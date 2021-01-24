import React, { useEffect, useState, useRef } from 'react';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls';
import _ from 'lodash';
import Select from 'react-select';
import { Link } from '@reach/router';
import { Button, Checkbox } from 'semantic-ui-react';
import { HuePicker } from 'react-color';

import styles from './index.module.scss';
import Service from '../../libs/service';
import Renderer from '../../libs/renderer';
import { COLOR } from '../../libs/constant';

const RenderSpacePage = (props) => {
  const mount = useRef(null);
  const control = useRef(null);

  const [renderingConfig, setRenderingConfig] = useState({
    group: null,
    scene: null,
    camera: null,
    renderer: null,
    rotating: true,
  });
  const [spaceId, setSpaceId] = useState();
  const [uiData, setUiData] = useState({
    spaceIdOptions: [],
  });
  const [spaceTime, setSpaceTime] = useState();

  const { spaceId: defaultSpaceId } = props;

  const fetchSpaceIds = async () => {
    const spaces = await Service.getSpaces();

    const spaceIdOptions = spaces.map((space) => {
      const spaceId = space.space_id;
      return { value: spaceId, label: spaceId };
    });
    setUiData({ ...uiData, spaceIdOptions });
  };

  const onChangeSpaceId = async (option) => {
    const spaceId = option.value;
    const spaceTime = await Service.getSpaceTime(spaceId);

    setSpaceTime(spaceTime);
  };

  // componentDidMount, will load once only when page start
  useEffect(() => {
    fetchSpaceIds();

    const { clientWidth: width, clientHeight: height } = mount.current;

    const { group, scene, camera, renderer } = Renderer.init(width, height);
    setRenderingConfig({ ...renderingConfig, group, scene, camera, renderer });

    const handleResize = () => {
      const { clientWidth: width, clientHeight: height } = mount.current;
      renderer.setSize(width, height);
      camera.aspect = width / height;
      camera.updateProjectionMatrix();
      renderer.render(scene, camera);
    };

    new OrbitControls(camera, renderer.domElement);
    mount.current.appendChild(renderer.domElement);
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, []);

  useEffect(() => {
    if (defaultSpaceId) {
      onChangeSpaceId({ value: spaceId });
    }
  }, [defaultSpaceId]);

  // called for every button clicks to update how the UI should render
  useEffect(() => {
    const { group, scene, camera, renderer, rotating } = renderingConfig;
    if (!(scene && camera && renderer && group)) return;

    const frameId = _.get(control, 'current.frameId');
    if (frameId) {
      cancelAnimationFrame(frameId);
    }

    const animate = () => {
      if (rotating) {
        group.rotation.y += 0.005;
      }
      const frameId = requestAnimationFrame(animate);
      control.current = { frameId };
      renderer.render(scene, camera);
    };

    animate();
  }, [renderingConfig]);

  return (
    <div className={styles.renderSpacePage}>
      <div className={styles.panel}>
        <div>
          <Button
            toggle
            active={renderingConfig.rotating}
            onClick={() =>
              setRenderingConfig({
                ...renderingConfig,
                rotating: !renderingConfig.rotating,
              })
            }
          >
            Rotate
          </Button>
        </div>
        <hr />
        <div>
          <span>Space ID:</span>
          <Select
            defaultValue={{
              value: defaultSpaceId,
              label: defaultSpaceId,
            }}
            onChange={onChangeSpaceId}
            options={uiData.spaceIdOptions}
            isSearchable={true}
          />
        </div>

        <hr />
      </div>
      <div className={styles.vis} ref={mount} />
    </div>
  );
};

export default RenderSpacePage;

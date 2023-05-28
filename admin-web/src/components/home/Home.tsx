import React, { useEffect, useState } from 'react';
import type { MenuProps } from 'antd';
import { Breadcrumb, Layout, Menu, theme } from 'antd';
import { ItemType } from 'antd/es/menu/hooks/useItems';
import {setItem, getItem, deleteItem} from '../../common/LocalStorageUtil';
import { ActiveWorkers } from '../workers/ActiveWorkers';
import { ActivityJobsView } from '../activities/ActivityJobsView';

const { Header, Content, Sider } = Layout;

export const Frame = () => {
  const [headerItems, setHeaderItems] = useState<ItemType[]>([{ key: 'activeWorkers', label: 'Active Workers' }, {key: 'activities', label: 'Activities'}]);
  const {
    token: { colorBgContainer },
  } = theme.useToken();
  const [headerNavKey, setHeaderNavKey] = useState<string>('activeWorkers');

  const displayPageSettings: Record<string, JSX.Element> = {
    ['activeWorkers']: <ActiveWorkers show={true}/>,
    ['activities']: <ActivityJobsView show={true}/>,
  };

  return (
    <Layout>
      <Header className="header">
        <div className="logo" />
        <Menu theme="dark" mode="horizontal" defaultSelectedKeys={[headerNavKey]} items={headerItems} onClick={evt => {
          setHeaderNavKey(evt.key);
        }} />
      </Header>
      <Layout>
        <Layout style={{ padding: '0 1vw 1vw' }}>
          <Breadcrumb style={{ margin: '1vw 0' }}>
            Location: {(headerItems.find(item => item!.key == headerNavKey) as {label: string}).label}&nbsp;
            <Breadcrumb.Item>/{}</Breadcrumb.Item>
          </Breadcrumb>
          <Content
            style={{
              padding: 24,
              margin: 0,
              minHeight: 280,
              background: colorBgContainer,
            }}
          >
            {displayPageSettings[headerNavKey]}
          </Content>
        </Layout>
      </Layout>
    </Layout>
  );
};
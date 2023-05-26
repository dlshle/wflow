import React from 'react';
import {Collapse} from 'antd';
const {Panel} = Collapse;

export const CollapsableModel = (props: React.PropsWithChildren<{title: string}>) => {
    return (
        <Collapse>
            <Panel header={props.title} key="1">
                {props.children}
            </Panel>
        </Collapse>
    );
};
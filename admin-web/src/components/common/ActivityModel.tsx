import React from 'react';
import {Activity} from '../../lib/types';
import {Card} from 'antd';
import { JobDispatchView } from './JobDispatchView';

export type ActivityModelProps = {
    activity: Activity;
};

export const ActivityView = (props: ActivityModelProps) => {
    const {activity} = props;
    return (
        <>
            <p>Activity ID: {activity.id}</p>
            <p>Activity Name: {activity.name}</p>
            <p>Activity Description: {activity.description}</p>
            <JobDispatchView activity={activity}/>
        </>
    );
};

export const ActivityCard = (props: ActivityModelProps) => {
    const {activity} = props;
    return (
        <Card title={activity.name}>
            <ActivityView activity={activity}/>
        </Card>
    );
};

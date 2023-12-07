import React, {useState} from 'react';
import {Activity, Worker} from '../../lib/types';
import {Card} from 'antd';
import { JobDispatchView } from './JobDispatchView';
import { RemoteView } from './RemoteView';
import { adminClient } from '../../lib/api/AdminClient';
import { WorkerView } from './WorkerModel';

export type ActivityModelProps = {
    activity: Activity;
    showAssociatedWorkers?: boolean;
};

const WorkersView = (props: {workers: Worker[]}) => {
    return <>
        {props.workers.map((worker) => <WorkerView worker={worker}/>)}
    </>
};

export const ActivityView = (props: ActivityModelProps) => {
    const {activity} = props;
    const [associatedWorkers, setAssociatedWorkers] = useState<Worker[]>([]);
    return (
        <>
            <p>Activity ID: {activity.id}</p>
            <p>Activity Name: {activity.name}</p>
            <p>Activity Description: {activity.description?? 'none'}</p>
            <JobDispatchView activity={activity} associatedWorkers={associatedWorkers}/>
            {props.showAssociatedWorkers && <RemoteView errorDisappearTimeout={5500} asyncViewRenderer={() => adminClient.getWorkersByActivityID(activity.id).then(workers => {
                setAssociatedWorkers(workers);
                return <WorkersView workers={workers}/>;
            })} failureMessage={`failed to fetch workers for activity ${activity.id}`} />}
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

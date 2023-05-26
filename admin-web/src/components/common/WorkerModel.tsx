import React from 'react';
import { Worker } from '../../lib/types';
import { Card, Collapse, Divider } from 'antd';
import { ActivityView } from './ActivityModel';
const { Panel } = Collapse;

export type WorkerModelProps = {
    worker: Worker;
};

const WorkerSystemStatView = (props: WorkerModelProps) => {
    const { worker } = props;
    if (!worker.system_stat) {
        return <div></div>
    }
    const systemStat = worker.system_stat;
    return (
        <Collapse>
            <Panel header="Woker System Stat" key="1">
                <p>CPU Count: {systemStat.cpu_count}</p>
                <p>CPU Usage: {systemStat.cpu_usage}</p>
                <p>Aval. Mem: {systemStat.available_memory_in_bytes}</p>
                <p>Totl. Mem: {systemStat.total_memory_in_bytes}</p>
                <p>Mem Usage: {systemStat.available_memory_in_bytes / systemStat.total_memory_in_bytes}</p>
            </Panel>
        </Collapse>
    );
};

const WorkerActiveJobsView = (props: WorkerModelProps) => {
    const { worker } = props;
    if (!worker.active_jobs?.length) {
        return <div></div>
    }
    const activeJobs = worker.active_jobs;
    return (
        <Collapse style={{marginTop: '1%', marginBottom: '1%'}}>
            <Panel header="Active Jobs" key="1">
                {activeJobs.map((job) => <p>{job}</p>)}
            </Panel>
        </Collapse>
    );
};

const WorkerSupportedActivitiesView = (props: WorkerModelProps) => {
    const { worker } = props;
    if (!worker.supported_activities?.length) {
        return <div></div>
    }
    const supportedActivities = worker.supported_activities;
    return (
        <Collapse style={{marginTop: '1%', marginBottom: '1%'}}>
            <Panel header="Active Activities" key="1">
                {supportedActivities.map((activity) => <><ActivityView activity={activity} /><Divider /></>)}
            </Panel>
        </Collapse>
    );
};

export const WorkerView = (props: WorkerModelProps) => {
    const { worker } = props;
    return (
        <>
            <p>Worker ID: {worker.id}</p>
            <p>Connected Worker: {worker.connected_server?? 'unknwon'}</p>
            <p>Worker Status: {worker.worker_status.toString()}</p>
            <p>Created At: {new Date(worker.created_at_seconds * 1000).toString()}</p>
            {worker.system_stat && <WorkerSystemStatView worker={worker} />}
            {worker.active_jobs?.length && <WorkerActiveJobsView worker={worker} />}
            {worker.supported_activities?.length && <WorkerSupportedActivitiesView worker={worker} />}
        </>
    );
};

export const WorkerCard = (props: WorkerModelProps) => {
    const { worker } = props;
    return (
        <Card title={'Worker ' + worker.id}>
            <WorkerView worker={worker} />
        </Card>
    );
};
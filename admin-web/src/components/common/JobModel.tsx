import React, { useState } from 'react';
import { JobLog, JobReport, WrappedLogs } from '../../lib/types';
import { Button, Card, Modal, Space } from 'antd';
import { CollapsableModel } from './CollapsableModel';
import { RemoteView } from './RemoteView';
import { adminClient } from '../../lib/api/AdminClient';

export type JobModelProps = {
    jobReport: JobReport;
};

export const JobLogView = (props: {jobLog: JobLog[], jobID: string}) => {
    const singleLogView = (log: JobLog) => {
        return (<>
            <Card size="small" title={new Date(log.timestamp * 1000).toISOString()}>
                {[...Object.keys(log.contexts)].map(key => <p>{key}:{(log.contexts as any)[key]}</p>)}
                <p>[{log.level}]: {log.message}</p>
            </Card>
            </>)
    };
    return (
        <Space direction="vertical" size={16}>
        {props.jobLog.map(log => singleLogView(log))}
        </Space>
    );
};

export const JobView = (props: JobModelProps) => {
    const jobReport = props.jobReport;
    const job = jobReport.job;
    const [showLogs, setShowLogs] = useState<boolean>(false);
    const logsModal = <Modal title="Job Logs" open={showLogs} onCancel={() => setShowLogs(false)} onOk={() => setShowLogs(false)}>
        <RemoteView asyncViewRenderer={() => adminClient.getJobLogs(job.id).then(logs => <JobLogView jobID={job.id} jobLog={logs}/>)}/>
        </Modal>;
    return (
        <>
            <p>ID: {job.id}</p>
            <p>Activity ID: {job.activity_id}</p>
            <p>Param: <CollapsableModel title="Job Param"><div>{job.param? atob(job.param as any): 'nil'}</div></CollapsableModel></p>
            <p>Status: {jobReport.status}</p>
            <p>Dispatched At: {!!job.dispatch_time_in_seconds? new Date(job.dispatch_time_in_seconds * 1000).toISOString(): 'unknown'}</p>
            <p>Started At: {!!jobReport.job_started_time_seconds? new Date(jobReport.job_started_time_seconds * 1000).toISOString(): 'unknown'}</p>
            <p>Eexecuting Worker: {jobReport.worker_id}</p>
            { jobReport.result && <CollapsableModel title="Job Result"><div>{jobReport.result? atob(jobReport.result as any): 'nil'}</div></CollapsableModel> }
            { jobReport.failure_reason && <CollapsableModel title="Job Failure Reason"><div>{jobReport.failure_reason}</div></CollapsableModel> }
            <br/>
            <Button type="primary" onClick={() => {
                setShowLogs(true);
            }}>View Logs</Button>
            {logsModal}
        </>
    );
};

export const JobCard = (props: JobModelProps) => {
    const jobReport = props.jobReport;
    const job = jobReport.job;
    return (
        <Card title={'Job ' + job.id}>
            <JobView jobReport={jobReport} />
        </Card>
    );
};
import React from 'react';
import { JobReport } from '../../lib/types';
import { uint8ToBase64 } from '../../common/utils';
import { Card } from 'antd';
import {CollapsableModel} from './CollapsableModel';

export type JobModelProps = {
    jobReport: JobReport;
};

export const JobModelView = (props: JobModelProps) => {
    const jobReport = props.jobReport;
    const job = jobReport.job;
    return (
        <Card title={'Job ' + job.id}>
            <p>ID: {job.id}</p>
            <p>Activity ID: {job.activity_id}</p>
            <p>Param: <CollapsableModel title="Job Param"><div>{uint8ToBase64(job.param)}</div></CollapsableModel></p>
            <p>Status: {jobReport.status}</p>
            <p>Activity ID: {job.dispatch_time_in_seconds}</p>
            <p>Dispatched At: {job.dispatch_time_in_seconds? new Date(job.dispatch_time_in_seconds * 1000): 'not dispatched'}</p>
            <p>Started At: {jobReport.job_started_time_seconds? new Date(jobReport.job_started_time_seconds * 1000): 'unstarted'}</p>
            <p>Eexecuting Worker: {jobReport.worker_id}</p>
            {jobReport.result && <CollapsableModel title="Job Result"><div>{uint8ToBase64(jobReport.result)}</div></CollapsableModel>}
            {jobReport.failure_reason && <CollapsableModel title="Job Failure Reason"><div>{jobReport.failure_reason}</div></CollapsableModel>}
        </Card>
    );
};
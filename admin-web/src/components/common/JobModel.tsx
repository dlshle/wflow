import React from 'react';
import { JobReport } from '../../lib/types';
import { uint8ToBase64 } from '../../common/utils';
import { Card } from 'antd';
import { CollapsableModel } from './CollapsableModel';

export type JobModelProps = {
    jobReport: JobReport;
};

export const JobView = (props: JobModelProps) => {
    const jobReport = props.jobReport;
    const job = jobReport.job;
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
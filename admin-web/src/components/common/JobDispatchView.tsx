import React, {useState} from 'react';
import { Activity, JobReport, Worker } from '../../lib/types';
import { Alert, Button, Card, Select } from 'antd';
import {CollapsableModel} from './CollapsableModel';
import TextArea from 'antd/es/input/TextArea';
import { adminClient } from '../../lib/api/AdminClient';

export type JobDispatchProps = {
    activity: Activity;
    associatedWorkers?: Worker[];
};

export const JobDispatchView = (props: JobDispatchProps) => {
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [param, setParam] = useState<string>('');
    const [popupMessage, setPopupMessage] = useState<string>('');
    const [popupType, setPopupType] = useState<'success' | 'info' | 'warning' | 'error'>('info');
    const [selectedWorkerID, setSelectedWorkerID] = useState<string>('');
    const activity = props.activity;
    const scheduleTitle = 'Job Schedule for ' + activity.id;
    return (
            <CollapsableModel title={scheduleTitle}>
            <>
            {popupMessage && <Alert message={popupMessage} type={popupType} showIcon closable onClose={() => setPopupMessage('')}/>}
            Job Parameter(in base64):
            <br/>
            <TextArea rows={4} onChange={evt => setParam(evt.target.value)}/>
            <br/>
            {!!props.associatedWorkers?.length && <Select
                defaultValue=''
                style={{ width: 120 }}
                onChange={value => setSelectedWorkerID(value)}
                options={[{label: 'Auto Dispatch', value: ''}, ...props.associatedWorkers.map(w => ({label: w.id, value: `${w.id}(${w.worker_ip?? 'unknown'})`}))]}
            />}
            <br/>
            <Button type="primary" style={{marginTop: '10px'}} loading={isLoading} onClick={() => {
                setIsLoading(true);
                adminClient.dispatchJob(activity.id, param, selectedWorkerID).then((jobReport: JobReport) => {
                    setPopupMessage(`Job ${jobReport.job.id} dispatched successfully, please go to activity page to check job details`);
                    setPopupType('success');
                }).catch(err => {
                    setPopupMessage('Failed to dispatch job due to ' + err);
                    setPopupType('error');
                }).finally(() => {
                    setIsLoading(false);
                });
            }}>Dispatch</Button>
            </>
            </CollapsableModel>
    );
};
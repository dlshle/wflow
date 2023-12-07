import React, {useState, useEffect} from 'react';
import {Modal, Table} from 'antd';
import { ActivityWithJobIDs } from '../../lib/types';
import { adminClient } from '../../lib/api/AdminClient';
import { ActivityView } from '../common/ActivityModel';
import { CollapsableModel, JobView } from '../common';
import { RemoteView } from '../common/RemoteView';

export type ActivityJobsViewProps = {
    show: boolean;
};

export const ActivityJobsView: React.FC<ActivityJobsViewProps> = (props) => {
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [activitiesWithJobIDs, setActivitiesWithJobIDs] = useState<ActivityWithJobIDs[]>([]);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [isJobModalOpen, setIsJobModalOpen] = useState(false);
    const [jobID, setJobID] = useState<string>('');
    const [selectedActivity, setSelectedActivity] = useState<ActivityWithJobIDs | null>(null);
    const showModal = () => {
        setIsModalOpen(true);
    };

    const handleOk = () => {
        setIsModalOpen(false);
    };

    const handleCancel = () => {
        setIsModalOpen(false);
    };

    const remoteJobModal = <Modal title="Job Description" open={isJobModalOpen} onOk={() => setIsJobModalOpen(false)} onCancel={() => setIsJobModalOpen(false)}> 
    {jobID && <RemoteView asyncViewRenderer={() => adminClient.getJob(jobID).then(job => <JobView jobReport={job}/>)} failureMessage={`failed to fetch job detail for ${jobID}`} />}
    </Modal>

    const activityModal = <Modal title="Activity Detail" open={isModalOpen} onOk={handleOk} onCancel={handleCancel}>
        {selectedActivity && <>
        <ActivityView activity={selectedActivity.activity} showAssociatedWorkers={true}/>
        <CollapsableModel title="Schedueled Jobs"> 
            {(selectedActivity.job_ids?? []).map((jobID) => <><a onClick={() => {
                setJobID(jobID);
                setIsJobModalOpen(true);
            }}>{jobID}</a><br/></>)}
        </CollapsableModel>
        </>}
    </Modal>

    useEffect(() => {
        if (props.show) {
            adminClient.getActivities().then((activities) => {
                setActivitiesWithJobIDs(activities);
            }).catch(err => {
                alert('failed to get activities ' + err);
            }).finally(() => {
                setIsLoading(false);
            });
        }
    }, [props.show]);
    const columns = [
        {
            title: 'ID',
            dataIndex: ['activity', 'id'],
            key: 'id',
            render: (val: string) => <a onClick={() => {
                setSelectedActivity(activitiesWithJobIDs.find(activity => activity.activity.id === val)?? null);
                showModal();
            }}>{val}</a>,
        },
        {
            title: 'Name',
            dataIndex: ['activity', 'name'],
            key: 'name',
        },
        {
            title: 'Description',
            key: ['activity', 'description'],
            dataIndex: 'description',
            render: (val: string) => val?? 'Unknown',
        },
        {
            title: 'Number of Jobs',
            dataIndex: 'job_ids',
            key: 'job_ids',
            render: (val: any) => val?.length?? 0,
        },
    ];
    // @ts-ignore
    return isLoading ? <p>'Loading'</p> : (<><Table columns={columns} dataSource={activitiesWithJobIDs} />{activityModal}{remoteJobModal}</>)
};
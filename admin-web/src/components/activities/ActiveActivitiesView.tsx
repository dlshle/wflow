import React, { useEffect, useState } from 'react';
import { Table, Modal } from 'antd';
import { Activity } from '../../lib/types';
import { adminClient } from '../../lib/api/AdminClient';
import { ActivityView } from '../common';

type ActiveActivitiesProps = {
    show: boolean;
};

export const ActiveActivities = (props: ActiveActivitiesProps) => {
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [activeActivities, setActiveActivities] = useState<Activity[]>([]);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [selectedActivity, setSelectedActivity] = useState<Activity | null>(null);
    const showModal = () => {
        setIsModalOpen(true);
    };

    const handleOk = () => {
        setIsModalOpen(false);
    };

    const handleCancel = () => {
        setIsModalOpen(false);
    };

    const activityModal = <Modal title="Activity Detail" open={isModalOpen} onOk={handleOk} onCancel={handleCancel}>
        {selectedActivity &&  <ActivityView activity={selectedActivity} />}
    </Modal>

    useEffect(() => {
        if (props.show) {
            adminClient.getActiveActivities().then((activities) => {
                setActiveActivities(activities);
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
            dataIndex: 'id',
            key: 'id',
            render: (id: string) => <a onClick={() => {
                setSelectedActivity(activeActivities.find(activity => activity.id === id)?? null);
                showModal();
            }}>{id}</a>,
        },
        {
            title: 'Name',
            dataIndex: 'name',
            key: 'name',
        },
        {
            title: 'Description',
            dataIndex: 'description',
            key: 'description',
            render: (val: string) => val?? 'N/A',
        }
    ];

    return isLoading ? <p>'Loading'</p> : (<><Table columns={columns} dataSource={activeActivities} />{activityModal}</>)
};
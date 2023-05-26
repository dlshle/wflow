import React, { useEffect, useState } from 'react';
import { Table, Modal } from 'antd';
import { Worker } from '../../lib/types';
import { adminClient } from '../../lib/api/AdminClient';
import { WorkerView } from '../common';
import {RemoteWorkerViewV2} from './RemoteWorkerView';

type ActiveWorkersProps = {
    show: boolean;
};

export const ActiveWorkers = (props: ActiveWorkersProps) => {
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [activeWorkers, setActiveWorkers] = useState<Worker[]>([]);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [selectedWorker, setSelectedWorker] = useState<Worker | null>(null);
    const [selectedWorkerID, setSelectedWorkerID] = useState<string>('');
    const showModal = () => {
        setIsModalOpen(true);
    };

    const handleOk = () => {
        setIsModalOpen(false);
    };

    const handleCancel = () => {
        setIsModalOpen(false);
    };

    // {selectedWorker && <WorkerView worker={selectedWorker} />}
    const workerModal = <Modal title="Worker Detail" open={isModalOpen} onOk={handleOk} onCancel={handleCancel}>
        {selectedWorkerID && <RemoteWorkerViewV2 workerId={selectedWorkerID} />}
    </Modal>

    useEffect(() => {
        if (props.show) {
            adminClient.getActiveWorkers().then((workers) => {
                const fetchedWorkers = [...Object.values(workers.workers ?? {})];
                setActiveWorkers(fetchedWorkers);
            }).catch(err => {
                alert('failed to get workers ' + err);
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
            render: (val: string) => <a onClick={() => {
                setSelectedWorkerID(val);
                showModal();
            }}>{val}</a>,
        },
        {
            title: 'Number of Supported Activities',
            dataIndex: 'supported_activities',
            key: 'supported_activities',
            render: (val: any) => val.length,
        },
        {
            title: 'Connected Server',
            dataIndex: 'connected_server',
            key: 'connected_server',
        },
        {
            title: 'Worker Status',
            key: 'worker_status',
            dataIndex: 'worker_status',
            // render: (val: number) => val === -1? '未知': ['XPrinter', 'GPrinter', 'SPRT', 'UCP'][val],
        }
    ];

    return isLoading ? <p>'Loading'</p> : (<><Table columns={columns} dataSource={activeWorkers} />{workerModal}</>)
};
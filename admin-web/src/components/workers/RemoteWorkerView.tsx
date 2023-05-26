import React, { useState } from 'react';
import { WorkerView } from '../common/WorkerModel';
import { adminClient } from '../../lib/api/AdminClient';
import { Worker } from '../../lib/types';
import {RemoteView} from '../common/RemoteView';

export type RemoteWorkerProps = {
    workerId: string;
};

export const RemoteWorkerViewV2 = (props: RemoteWorkerProps) => <RemoteView asyncViewRenderer={() => adminClient.getWorker(props.workerId).then(worker => <WorkerView worker={worker}/>)} failureMessage={'failed to fetch worker ' + props.workerId} />;
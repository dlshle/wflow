import React, { useState } from 'react';

export type RemoteViewProps = {
    asyncViewRenderer: () => Promise<JSX.Element>;
    failureMessage?: string;
};

export function RemoteView(props: RemoteViewProps) {
    const [view, setView] = useState<JSX.Element | null>(null);
    const [failed, setFailed] = useState<boolean>(false);
    const [fulfilled, setFulfilled] = useState<boolean>(false);
    React.useEffect(() => {
        props.asyncViewRenderer().then(producedView => {
            setView(producedView);
        }).catch(err => {
            alert('failed to get worker ' + err);
            setFailed(true);
        }).finally(() => setFulfilled(true));
    }, [props.asyncViewRenderer]);
    return (
        <div>
            {!fulfilled ? <div>loading...</div> : failed ? <p style={{ color: 'red' }}>{props.failureMessage ?? 'failed to render view'}</p> : view}
        </div>
    );
};
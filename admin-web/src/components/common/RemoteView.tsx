import React, { useState } from 'react';

export type RemoteViewProps = {
    asyncViewRenderer: () => Promise<JSX.Element>;
    failureMessage?: string;
    errorDisappearTimeout?: number;
};

export function RemoteView(props: RemoteViewProps) {
    const [view, setView] = useState<JSX.Element | null>(null);
    const [failed, setFailed] = useState<boolean>(false);
    const [fulfilled, setFulfilled] = useState<boolean>(false);
    const [hideErrMsg, setHideErrMsg] = useState<boolean>(false);
    const errorView = <div hidden={hideErrMsg} style={{border: '1px solid red'}}><label style={{ color: 'red' }}>{props.failureMessage ?? 'failed to render view'}</label></div>;
    React.useEffect(() => {
        props.asyncViewRenderer().then(producedView => {
            setView(producedView);
        }).catch(err => {
            setFailed(true);
            props.errorDisappearTimeout && setTimeout(() => setHideErrMsg(true), props.errorDisappearTimeout);
        }).finally(() => setFulfilled(true));
    }, [props.asyncViewRenderer]);
    return (
        <div>
            {!fulfilled ? <div>loading...</div> : failed ? errorView: view}
        </div>
    );
};
import React, { useRef } from 'react';
import { Modal } from 'antd';
import TextMediaPanel from './TextMediaPanel.jsx';

export default function TextMediaDialog(props) {
  const requestClose=useRef(null);
  return <Modal open={Boolean(props.book)} title={null} onCancel={()=>requestClose.current?.()}
    footer={null} width="min(1180px, calc(100vw - 40px))" closable={false}
    className="batch-factory-media-library-modal is-text" destroyOnClose>
    {props.book && <TextMediaPanel {...props} closeGuard={requestClose}/>}
  </Modal>;
}

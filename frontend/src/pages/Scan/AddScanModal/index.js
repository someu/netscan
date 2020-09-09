import React, { useEffect } from "react";
import { Table, message, Button, Input, Modal, Form } from "antd";
import axios from "axios";
import { useRequest } from "ahooks";

function AddScanModal(props) {
  const { visible, onCancel } = props;
  const { loading: adding, run: addScan } = useRequest(
    (values) => axios.post("/api/scan", values),
    {
      manual: true,
    }
  );

  const [form] = Form.useForm();

  const onModalOk = async () => {
    try {
      const values = await form.validateFields();
      const formdata = new FormData();
      formdata.append("target", values.target);
      addScan(formdata)
        .then(() => {
          message.success("Add scan successed");
          onCancel();
        })
        .catch(() => {
          message.error("Add scan failed");
        });
    } catch {}
  };

  return (
    <Modal
      className="add-scan-modal"
      title="Add Scan"
      visible={visible}
      onCancel={onCancel}
      onOk={onModalOk}
      loading={adding}
    >
      <Form className="form" form={form} name="add-scan-from">
        <Form.Item
          name="target"
          label="Target"
          rules={[
            {
              required: true,
              message: "target is required",
            },
          ]}
          colon={false}
        >
          <Input.TextArea
            className="target"
            placeholder="input scan target, multiply targets input line-by-line"
            rows={4}
          />
        </Form.Item>
      </Form>
    </Modal>
  );
}

export default AddScanModal;

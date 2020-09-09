import React, { useState, useEffect } from "react";
import { Table, message, Button, Input } from "antd";
import { useRequest } from "ahooks";
import axios from "axios";
import _ from "lodash";
import { PlusOutlined } from "@ant-design/icons";
import qs from "query-string";
import AddScanModal from "./AddScanModal";
import "./index.less";

const columns = [
  {
    title: "Target",
    dataIndex: "Target",
  },
  {
    title: "Status",
    dataIndex: "Status",
  },
  {
    title: "StartAt",
    dataIndex: "StartAt",
  },
  {
    title: "FinishAt",
    dataIndex: "FinishAt",
  },
  {
    title: "Operator",
    className: "scan-opers",
    render: () => {
      return (
        <div>
          <a>stop</a>
          <a>cancle</a>
          <a>remove</a>
        </div>
      );
    },
  },
];

export default function () {
  const [data, setData] = useState({});
  const [filter, setFilter] = useState({
    pageSize: 10,
    current: 1,
    search: "",
  });
  const [addScanModalVisible, setAddScanModalVisible] = useState(false);
  const { run, loading, cancel } = useRequest(
    () => {
      const url = `/api/scan?${qs.stringify(filter)}`;
      return axios.get(url);
    },
    {
      paginated: true,
      manual: true,
      onSuccess: (res) => {
        console.log(res);
        setData(_.get(res, "data"));
      },
      onError: () => {
        message.error("Request scan list failed");
      },
    }
  );

  const onTableChange = ({ current, pageSize }) => {
    setFilter(Object.assign({}, filter, { current, pageSize }));
  };

  useEffect(() => {
    run();
    return cancel;
  }, []);
  // useEffect(() => run(), [filter]);

  return (
    <div className="scan">
      <h2>Scan</h2>
      <div className="oper">
        <Button type="primary" onClick={() => setAddScanModalVisible(true)}>
          <PlusOutlined />
          Add Scan
        </Button>
        <Input.Search
          className="search"
          placeholder="search scans by target"
          onSearch={(search) =>
            setFilter(Object.assign({}, filter, { search }))
          }
        />
      </div>
      <Table
        rowKey="_id"
        dataSource={_.get(data, "List", [])}
        columns={columns}
        loading={loading}
        onChange={onTableChange}
        pagination={{
          current: filter.current,
          pageSize: filter.pageSize,
          total: _.get(data, "Total", 0),
          showQuickJumper: true,
          showSizeChanger: true,
        }}
      />
      <AddScanModal
        visible={addScanModalVisible}
        onCancel={() => setAddScanModalVisible(false)}
      />
    </div>
  );
}

import React, { useState, useEffect } from "react";
import { Table, message } from "antd";
import { useRequest } from "ahooks";
import axios from "axios";
import _ from "lodash";

const columns = [
  {
    title: "Address",
    dataIndex: "address",
  },
];

export default function () {
  const [data, setData] = useState({});
  const [filter, setFilter] = useState({
    pageSize: 10,
    current: 1,
    search: "",
  });
  const { run, loading } = useRequest(
    () => {
      const url = `/api/asset?current=${filter}&pageSize=${pageSize}`;
      return axios.get(url);
    },
    {
      paginated: true,
      manual: true,
      onSuccess: (res) => {
        setData(_.get(res, "data"));
      },
      onError: () => {
        message.error("Request asset list failed");
      },
    }
  );

  const onTableChange = ({ current, pageSize }) => {
    setFilter(Object.assign({}, filter, { current, pageSize }));
  };

  useEffect(() => {
    run();
  }, []);

  return (
    <div>
      <h2>Asset</h2>
      <Table
        dataSource={_.get(data, "list", [])}
        columns={columns}
        loading={loading}
        onChange={onTableChange}
        pagination={{
          current: filter.current,
          pageSize: filter.pageSize,
          total: _.get(data, "total", 0),
          showQuickJumper: true,
          showSizeChanger: true,
        }}
      />
    </div>
  );
}

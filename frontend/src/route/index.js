import React from "react";
import { Layout, Menu } from "antd";
import {
  BrowserRouter,
  Switch,
  Redirect,
  Route,
  NavLink,
} from "react-router-dom";
import Asset from "../pages/Asset";
import Config from "../pages/Config";
import Scan from "../pages/Scan";
import "./index.less";

function RootRoute() {
  const menus = [
    {
      to: "/",
      title: "Asset",
    },
    {
      to: "/scan",
      title: "Scan",
    },
    {
      to: "/config",
      title: "Config",
    },
  ];
  return (
    <BrowserRouter>
      <Layout id="app-layout">
        <Layout.Sider id="app-sider">
          <Menu className="menu">
            {menus.map((menu) => (
              <Menu.Item key={menu.to} title={menu.title}>
                <NavLink to={menu.to} exact>
                  {menu.title}
                </NavLink>
              </Menu.Item>
            ))}
          </Menu>
        </Layout.Sider>
        <Layout.Content id="app-content">
          <div className="inner">
            <Switch>
              <Route path="/" component={Asset} exact />
              <Route path="/scan" component={Scan} exact />
              <Route path="/config" component={Config} exact />
              <Route component={() => <Redirect to="/" />} />
            </Switch>
          </div>
        </Layout.Content>
      </Layout>
    </BrowserRouter>
  );
}

export default RootRoute;

import React, { Component } from "react";

export default class ErrorBoundary extends Component {
  state = {
    error: null,
  };
  componentDidCatch(error) {
    this.setState({ error });
  }
  render() {
    return this.state.error ? <div>{error}</div> : this.props.children;
  }
}

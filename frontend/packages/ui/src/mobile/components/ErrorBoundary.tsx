import React from 'react';
import { View, Text } from 'react-native';

export class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    console.error("ErrorBoundary caught error:", error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <View style={{ padding: 20, backgroundColor: 'red', margin: 10, borderRadius: 10 }}>
          <Text style={{ color: 'white', fontWeight: 'bold' }}>Error rendering item</Text>
          <Text style={{ color: 'white', fontSize: 10 }}>{String(this.state.error)}</Text>
        </View>
      );
    }
    return this.props.children;
  }
}

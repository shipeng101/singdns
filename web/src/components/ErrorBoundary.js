import React from 'react';
import {
  Box,
  Typography,
  Button,
  Container,
  Paper
} from '@mui/material';
import { Warning as WarningIcon } from '@mui/icons-material';

class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true };
  }

  componentDidCatch(error, errorInfo) {
    this.setState({
      error: error,
      errorInfo: errorInfo
    });
    // 这里可以添加错误日志上报逻辑
    console.error('Error:', error);
    console.error('Error Info:', errorInfo);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null, errorInfo: null });
  };

  render() {
    if (this.state.hasError) {
      return (
        <Container maxWidth="sm">
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              minHeight: '100vh',
              textAlign: 'center'
            }}
          >
            <Paper
              elevation={3}
              sx={{
                p: 4,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                gap: 2
              }}
            >
              <WarningIcon color="error" sx={{ fontSize: 64 }} />
              <Typography variant="h5" gutterBottom>
                出错了
              </Typography>
              <Typography variant="body1" color="text.secondary" paragraph>
                应用程序遇到了一个错误。请尝试刷新页面或联系管理员。
              </Typography>
              {process.env.NODE_ENV === 'development' && (
                <Box sx={{ mt: 2, textAlign: 'left', width: '100%' }}>
                  <Typography variant="subtitle2" color="error">
                    {this.state.error && this.state.error.toString()}
                  </Typography>
                  <Typography
                    variant="body2"
                    component="pre"
                    sx={{
                      mt: 1,
                      p: 2,
                      bgcolor: 'grey.100',
                      borderRadius: 1,
                      overflow: 'auto'
                    }}
                  >
                    {this.state.errorInfo && this.state.errorInfo.componentStack}
                  </Typography>
                </Box>
              )}
              <Box sx={{ mt: 2, display: 'flex', gap: 2 }}>
                <Button
                  variant="contained"
                  onClick={() => window.location.reload()}
                >
                  刷新页面
                </Button>
                <Button
                  variant="outlined"
                  onClick={this.handleReset}
                >
                  重试
                </Button>
              </Box>
            </Paper>
          </Box>
        </Container>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary; 
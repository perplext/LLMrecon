import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  CardHeader,
  CircularProgress,
  Container,
  Divider,
  FormControl,
  FormControlLabel,
  FormLabel,
  Radio,
  RadioGroup,
  TextField,
  Typography
} from '@mui/material';

// MFA verification component used during login
const MFAVerify = ({ onVerify, onCancel, availableMethods, defaultMethod }) => {
  const [selectedMethod, setSelectedMethod] = useState(defaultMethod || 'totp');
  const [verificationCode, setVerificationCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Set default method when component mounts
  useEffect(() => {
    if (defaultMethod && availableMethods.includes(defaultMethod)) {
      setSelectedMethod(defaultMethod);
    } else if (availableMethods && availableMethods.length > 0) {
      setSelectedMethod(availableMethods[0]);
    }
  }, [defaultMethod, availableMethods]);

  // Handle method change
  const handleMethodChange = (event) => {
    setSelectedMethod(event.target.value);
    setVerificationCode('');
    setError('');
  };

  // Handle verification code change
  const handleCodeChange = (event) => {
    setVerificationCode(event.target.value);
    if (error) setError('');
  };

  // Handle verification
  const handleVerify = async () => {
    // Validate input
    if (!verificationCode && selectedMethod !== 'webauthn') {
      setError('Please enter a verification code');
      return;
    }

    setLoading(true);
    setError('');

    try {
      // Call the appropriate verification endpoint based on the selected method
      let response;
      
      if (selectedMethod === 'totp') {
        response = await verifyTOTP();
      } else if (selectedMethod === 'backup') {
        response = await verifyBackupCode();
      } else if (selectedMethod === 'webauthn') {
        response = await verifyWebAuthn();
      } else if (selectedMethod === 'sms') {
        response = await verifySMS();
      }

      if (response && response.success) {
        // Call the onVerify callback to proceed with login
        if (onVerify) {
          onVerify();
        }
      } else {
        setError('Verification failed. Please try again.');
      }
    } catch (err) {
      setError(err.message || 'Verification failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  // Verify TOTP
  const verifyTOTP = async () => {
    const response = await fetch('/api/mfa/verify', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ 
        method: 'totp',
        code: verificationCode 
      }),
    });

    if (!response.ok) {
      throw new Error('Invalid verification code');
    }

    return await response.json();
  };

  // Verify backup code
  const verifyBackupCode = async () => {
    const response = await fetch('/api/mfa/verify', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ 
        method: 'backup',
        code: verificationCode 
      }),
    });

    if (!response.ok) {
      throw new Error('Invalid backup code');
    }

    return await response.json();
  };

  // Verify WebAuthn
  const verifyWebAuthn = async () => {
    // First, get the authentication options
    const beginResponse = await fetch('/api/mfa/webauthn/authenticate-begin', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!beginResponse.ok) {
      throw new Error('Failed to start WebAuthn authentication');
    }

    const options = await beginResponse.json();
    
    // In a real implementation, we would use the WebAuthn API here
    // For now, we'll just simulate the process
    console.log('WebAuthn authentication options:', options);
    
    // Simulate successful authentication
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    // Complete the authentication
    const completeResponse = await fetch('/api/mfa/webauthn/authenticate-complete', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ 
        assertionResponse: 'simulated-webauthn-response' 
      }),
    });

    if (!completeResponse.ok) {
      throw new Error('WebAuthn authentication failed');
    }

    return await completeResponse.json();
  };

  // Verify SMS
  const verifySMS = async () => {
    const response = await fetch('/api/mfa/verify', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ 
        method: 'sms',
        code: verificationCode 
      }),
    });

    if (!response.ok) {
      throw new Error('Invalid SMS code');
    }

    return await response.json();
  };

  // Request a new SMS code
  const requestNewSMSCode = async () => {
    setLoading(true);
    setError('');

    try {
      const response = await fetch('/api/mfa/sms/setup', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error('Failed to send new SMS code');
      }

      // Show success message
      setError('');
      alert('A new verification code has been sent to your phone.');
    } catch (err) {
      setError(err.message || 'Failed to send new SMS code');
    } finally {
      setLoading(false);
    }
  };

  // Render method selection
  const renderMethodSelection = () => {
    if (!availableMethods || availableMethods.length <= 1) {
      return null;
    }

    return (
      <Box sx={{ mb: 3 }}>
        <FormControl component="fieldset">
          <FormLabel component="legend">Verification Method</FormLabel>
          <RadioGroup
            aria-label="mfa-method"
            name="mfa-method"
            value={selectedMethod}
            onChange={handleMethodChange}
          >
            {availableMethods.includes('totp') && (
              <FormControlLabel
                value="totp"
                control={<Radio />}
                label="Authenticator App"
              />
            )}
            {availableMethods.includes('backup') && (
              <FormControlLabel
                value="backup"
                control={<Radio />}
                label="Backup Code"
              />
            )}
            {availableMethods.includes('webauthn') && (
              <FormControlLabel
                value="webauthn"
                control={<Radio />}
                label="Security Key"
              />
            )}
            {availableMethods.includes('sms') && (
              <FormControlLabel
                value="sms"
                control={<Radio />}
                label="SMS Code"
              />
            )}
          </RadioGroup>
        </FormControl>
      </Box>
    );
  };

  // Render verification form
  const renderVerificationForm = () => {
    if (selectedMethod === 'webauthn') {
      return (
        <Box sx={{ textAlign: 'center', my: 3 }}>
          <Typography variant="body1" paragraph>
            Please insert your security key and follow the prompts.
          </Typography>
          {loading && <CircularProgress sx={{ my: 2 }} />}
        </Box>
      );
    }

    return (
      <Box sx={{ my: 3 }}>
        <Typography variant="body1" paragraph>
          {selectedMethod === 'totp'
            ? 'Enter the code from your authenticator app.'
            : selectedMethod === 'backup'
            ? 'Enter one of your backup codes.'
            : 'Enter the verification code sent to your phone.'}
        </Typography>

        <TextField
          label="Verification Code"
          variant="outlined"
          fullWidth
          value={verificationCode}
          onChange={handleCodeChange}
          error={!!error}
          helperText={error}
          sx={{ mb: 2 }}
          autoFocus
        />

        {selectedMethod === 'sms' && (
          <Button
            variant="text"
            color="primary"
            onClick={requestNewSMSCode}
            disabled={loading}
            sx={{ mb: 2 }}
          >
            Send a new code
          </Button>
        )}
      </Box>
    );
  };

  return (
    <Container maxWidth="sm">
      <Card>
        <CardHeader
          title="Two-Factor Authentication"
          subheader="Verify your identity to continue"
        />
        <Divider />
        <CardContent>
          {renderMethodSelection()}
          {renderVerificationForm()}

          <Box sx={{ display: 'flex', justifyContent: 'space-between', mt: 2 }}>
            <Button
              variant="outlined"
              onClick={onCancel}
              disabled={loading}
            >
              Cancel
            </Button>
            <Button
              variant="contained"
              color="primary"
              onClick={handleVerify}
              disabled={loading || (!verificationCode && selectedMethod !== 'webauthn')}
            >
              {loading ? <CircularProgress size={24} /> : 'Verify'}
            </Button>
          </Box>
        </CardContent>
      </Card>
    </Container>
  );
};

export default MFAVerify;

import React, { useState, useEffect } from 'react';
import { 
  Box, 
  Button, 
  Card, 
  CardContent, 
  CardHeader, 
  CircularProgress, 
  Container, 
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Divider, 
  FormControl, 
  FormControlLabel, 
  FormLabel, 
  Grid, 
  Radio, 
  RadioGroup, 
  Step, 
  StepLabel, 
  Stepper, 
  TextField, 
  Typography 
} from '@mui/material';
import QRCode from 'qrcode.react';
import { useAuth } from '../../contexts/AuthContext';

// MFA setup component
const MFASetup = () => {
  const { user, updateUser } = useAuth();
  const [activeStep, setActiveStep] = useState(0);
  const [mfaMethod, setMfaMethod] = useState('totp');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [totpSecret, setTotpSecret] = useState('');
  const [qrCodeUrl, setQrCodeUrl] = useState('');
  const [verificationCode, setVerificationCode] = useState('');
  const [phoneNumber, setPhoneNumber] = useState('');
  const [backupCodes, setBackupCodes] = useState([]);
  const [showBackupCodes, setShowBackupCodes] = useState(false);
  const [setupComplete, setSetupComplete] = useState(false);

  // Steps for the MFA setup process
  const steps = ['Select Method', 'Setup', 'Verify'];

  // Check if MFA is already enabled
  useEffect(() => {
    if (user && user.mfaEnabled) {
      setSetupComplete(true);
    }
  }, [user]);

  // Handle method selection
  const handleMethodChange = (event) => {
    setMfaMethod(event.target.value);
  };

  // Handle next step
  const handleNext = async () => {
    if (activeStep === 0) {
      // Method selection step
      setActiveStep(1);
      await setupSelectedMethod();
    } else if (activeStep === 1) {
      // Setup step
      setActiveStep(2);
    } else if (activeStep === 2) {
      // Verification step
      await verifySelectedMethod();
    }
  };

  // Handle back step
  const handleBack = () => {
    setActiveStep((prevStep) => prevStep - 1);
  };

  // Setup the selected MFA method
  const setupSelectedMethod = async () => {
    setLoading(true);
    setError('');

    try {
      if (mfaMethod === 'totp') {
        await setupTOTP();
      } else if (mfaMethod === 'backup') {
        await setupBackupCodes();
      } else if (mfaMethod === 'webauthn') {
        await setupWebAuthn();
      } else if (mfaMethod === 'sms') {
        // SMS setup will be done in the next step
      }
    } catch (err) {
      setError(err.message || 'Failed to setup MFA method');
    } finally {
      setLoading(false);
    }
  };

  // Setup TOTP
  const setupTOTP = async () => {
    const response = await fetch('/api/mfa/totp/setup', {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error('Failed to setup TOTP');
    }

    const data = await response.json();
    setTotpSecret(data.secret);
    setQrCodeUrl(data.qr_code_url);
  };

  // Setup backup codes
  const setupBackupCodes = async () => {
    const response = await fetch('/api/mfa/backup-codes/generate', {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error('Failed to generate backup codes');
    }

    const data = await response.json();
    setBackupCodes(data.codes);
  };

  // Setup WebAuthn
  const setupWebAuthn = async () => {
    const response = await fetch('/api/mfa/webauthn/register-begin', {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error('Failed to setup WebAuthn');
    }

    const options = await response.json();
    
    // In a real implementation, we would use the WebAuthn API here
    // For now, we'll just simulate the process
    console.log('WebAuthn registration options:', options);
    
    // Simulate successful registration
    setTimeout(() => {
      // Move to verification step
      setActiveStep(2);
    }, 2000);
  };

  // Setup SMS
  const setupSMS = async () => {
    if (!phoneNumber) {
      setError('Phone number is required');
      return;
    }

    const response = await fetch('/api/mfa/sms/setup', {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ phone_number: phoneNumber }),
    });

    if (!response.ok) {
      throw new Error('Failed to setup SMS verification');
    }

    // SMS code has been sent to the user's phone
  };

  // Verify the selected MFA method
  const verifySelectedMethod = async () => {
    setLoading(true);
    setError('');

    try {
      if (mfaMethod === 'totp') {
        await verifyTOTP();
      } else if (mfaMethod === 'backup') {
        // Backup codes don't need verification
        await enableMFA();
      } else if (mfaMethod === 'webauthn') {
        await verifyWebAuthn();
      } else if (mfaMethod === 'sms') {
        await verifySMS();
      }
    } catch (err) {
      setError(err.message || 'Failed to verify MFA method');
    } finally {
      setLoading(false);
    }
  };

  // Verify TOTP
  const verifyTOTP = async () => {
    if (!verificationCode) {
      setError('Verification code is required');
      return;
    }

    const response = await fetch('/api/mfa/totp/verify', {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ code: verificationCode }),
    });

    if (!response.ok) {
      throw new Error('Invalid verification code');
    }

    // TOTP verification successful
    setSetupComplete(true);
    
    // Update user info
    if (updateUser) {
      updateUser({ ...user, mfaEnabled: true });
    }
  };

  // Verify WebAuthn
  const verifyWebAuthn = async () => {
    const response = await fetch('/api/mfa/webauthn/authenticate-begin', {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error('Failed to start WebAuthn authentication');
    }

    const options = await response.json();
    
    // In a real implementation, we would use the WebAuthn API here
    // For now, we'll just simulate the process
    console.log('WebAuthn authentication options:', options);
    
    // Simulate successful authentication
    setTimeout(() => {
      setSetupComplete(true);
      
      // Update user info
      if (updateUser) {
        updateUser({ ...user, mfaEnabled: true });
      }
    }, 2000);
  };

  // Verify SMS
  const verifySMS = async () => {
    if (!verificationCode) {
      setError('Verification code is required');
      return;
    }

    const response = await fetch('/api/mfa/sms/verify', {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ code: verificationCode }),
    });

    if (!response.ok) {
      throw new Error('Invalid verification code');
    }

    // SMS verification successful
    setSetupComplete(true);
    
    // Update user info
    if (updateUser) {
      updateUser({ ...user, mfaEnabled: true });
    }
  };

  // Enable MFA
  const enableMFA = async () => {
    const response = await fetch('/api/mfa/enable', {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ method: mfaMethod }),
    });

    if (!response.ok) {
      throw new Error('Failed to enable MFA');
    }

    // MFA enabled successfully
    setSetupComplete(true);
    
    // Update user info
    if (updateUser) {
      updateUser({ ...user, mfaEnabled: true });
    }
  };

  // Disable MFA
  const disableMFA = async () => {
    setLoading(true);
    setError('');

    try {
      const response = await fetch('/api/mfa/disable', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error('Failed to disable MFA');
      }

      // MFA disabled successfully
      setSetupComplete(false);
      
      // Update user info
      if (updateUser) {
        updateUser({ ...user, mfaEnabled: false });
      }
      
      // Reset state
      setActiveStep(0);
      setMfaMethod('totp');
      setTotpSecret('');
      setQrCodeUrl('');
      setVerificationCode('');
      setPhoneNumber('');
      setBackupCodes([]);
    } catch (err) {
      setError(err.message || 'Failed to disable MFA');
    } finally {
      setLoading(false);
    }
  };

  // Render method selection step
  const renderMethodSelection = () => {
    return (
      <Box sx={{ mt: 2 }}>
        <FormControl component="fieldset">
          <FormLabel component="legend">Select MFA Method</FormLabel>
          <RadioGroup
            aria-label="mfa-method"
            name="mfa-method"
            value={mfaMethod}
            onChange={handleMethodChange}
          >
            <FormControlLabel
              value="totp"
              control={<Radio />}
              label="Authenticator App (TOTP)"
            />
            <FormControlLabel
              value="backup"
              control={<Radio />}
              label="Backup Codes"
            />
            <FormControlLabel
              value="webauthn"
              control={<Radio />}
              label="Security Key (WebAuthn/FIDO2)"
            />
            <FormControlLabel
              value="sms"
              control={<Radio />}
              label="SMS Verification"
            />
          </RadioGroup>
        </FormControl>
      </Box>
    );
  };

  // Render setup step
  const renderSetup = () => {
    if (mfaMethod === 'totp') {
      return renderTOTPSetup();
    } else if (mfaMethod === 'backup') {
      return renderBackupCodesSetup();
    } else if (mfaMethod === 'webauthn') {
      return renderWebAuthnSetup();
    } else if (mfaMethod === 'sms') {
      return renderSMSSetup();
    }

    return null;
  };

  // Render TOTP setup
  const renderTOTPSetup = () => {
    return (
      <Box sx={{ mt: 2 }}>
        <Typography variant="h6" gutterBottom>
          Authenticator App Setup
        </Typography>
        <Typography variant="body1" paragraph>
          1. Install an authenticator app on your mobile device (like Google Authenticator, Authy, or Microsoft Authenticator).
        </Typography>
        <Typography variant="body1" paragraph>
          2. Scan the QR code below with your authenticator app.
        </Typography>
        <Typography variant="body1" paragraph>
          3. If you can't scan the QR code, enter this secret key manually: <strong>{totpSecret}</strong>
        </Typography>

        <Box sx={{ display: 'flex', justifyContent: 'center', my: 3 }}>
          {qrCodeUrl ? (
            <QRCode value={qrCodeUrl} size={200} />
          ) : (
            <CircularProgress />
          )}
        </Box>
      </Box>
    );
  };

  // Render backup codes setup
  const renderBackupCodesSetup = () => {
    return (
      <Box sx={{ mt: 2 }}>
        <Typography variant="h6" gutterBottom>
          Backup Codes
        </Typography>
        <Typography variant="body1" paragraph>
          Save these backup codes in a secure place. Each code can only be used once.
        </Typography>

        <Box sx={{ 
          border: '1px solid #ccc', 
          borderRadius: 1, 
          p: 2, 
          my: 2,
          maxHeight: '200px',
          overflowY: 'auto',
          backgroundColor: '#f5f5f5'
        }}>
          {backupCodes.length > 0 ? (
            <Grid container spacing={2}>
              {backupCodes.map((code, index) => (
                <Grid item xs={6} key={index}>
                  <Typography variant="body2" fontFamily="monospace">
                    {code}
                  </Typography>
                </Grid>
              ))}
            </Grid>
          ) : (
            <CircularProgress />
          )}
        </Box>

        <Button 
          variant="outlined" 
          onClick={() => {
            // In a real app, this would copy the codes to clipboard
            alert('Codes copied to clipboard');
          }}
        >
          Copy Codes
        </Button>
        <Button 
          variant="outlined" 
          sx={{ ml: 2 }}
          onClick={() => {
            // In a real app, this would download the codes as a text file
            alert('Codes downloaded');
          }}
        >
          Download Codes
        </Button>
      </Box>
    );
  };

  // Render WebAuthn setup
  const renderWebAuthnSetup = () => {
    return (
      <Box sx={{ mt: 2 }}>
        <Typography variant="h6" gutterBottom>
          Security Key Setup
        </Typography>
        <Typography variant="body1" paragraph>
          Connect your security key to your device and follow the prompts to register it.
        </Typography>

        <Box sx={{ display: 'flex', justifyContent: 'center', my: 3 }}>
          <CircularProgress />
        </Box>
      </Box>
    );
  };

  // Render SMS setup
  const renderSMSSetup = () => {
    return (
      <Box sx={{ mt: 2 }}>
        <Typography variant="h6" gutterBottom>
          SMS Verification Setup
        </Typography>
        <Typography variant="body1" paragraph>
          Enter your phone number to receive verification codes via SMS.
        </Typography>

        <TextField
          label="Phone Number"
          variant="outlined"
          fullWidth
          value={phoneNumber}
          onChange={(e) => setPhoneNumber(e.target.value)}
          placeholder="+1 (555) 123-4567"
          sx={{ mb: 2 }}
        />

        <Button 
          variant="contained" 
          color="primary"
          onClick={setupSMS}
          disabled={!phoneNumber || loading}
        >
          {loading ? <CircularProgress size={24} /> : 'Send Verification Code'}
        </Button>
      </Box>
    );
  };

  // Render verification step
  const renderVerification = () => {
    if (mfaMethod === 'backup') {
      return (
        <Box sx={{ mt: 2 }}>
          <Typography variant="body1" paragraph>
            Your backup codes have been generated. Keep them in a safe place.
          </Typography>
          <Button 
            variant="contained" 
            color="primary"
            onClick={() => setShowBackupCodes(true)}
          >
            Show Backup Codes Again
          </Button>
        </Box>
      );
    }

    return (
      <Box sx={{ mt: 2 }}>
        <Typography variant="h6" gutterBottom>
          Verify {mfaMethod === 'totp' ? 'Authenticator App' : 
                  mfaMethod === 'webauthn' ? 'Security Key' : 
                  'SMS Verification'}
        </Typography>
        
        {mfaMethod !== 'webauthn' && (
          <>
            <Typography variant="body1" paragraph>
              {mfaMethod === 'totp' 
                ? 'Enter the code from your authenticator app.'
                : 'Enter the verification code sent to your phone.'}
            </Typography>

            <TextField
              label="Verification Code"
              variant="outlined"
              fullWidth
              value={verificationCode}
              onChange={(e) => setVerificationCode(e.target.value)}
              sx={{ mb: 2 }}
            />
          </>
        )}

        {mfaMethod === 'webauthn' && (
          <Box sx={{ display: 'flex', justifyContent: 'center', my: 3 }}>
            <CircularProgress />
          </Box>
        )}
      </Box>
    );
  };

  // Render setup complete
  const renderSetupComplete = () => {
    return (
      <Box sx={{ mt: 2, textAlign: 'center' }}>
        <Typography variant="h6" gutterBottom color="success.main">
          Multi-Factor Authentication Enabled
        </Typography>
        <Typography variant="body1" paragraph>
          Your account is now protected with an additional layer of security.
        </Typography>
        <Button 
          variant="outlined" 
          color="error"
          onClick={disableMFA}
          sx={{ mt: 2 }}
        >
          Disable MFA
        </Button>
      </Box>
    );
  };

  return (
    <Container maxWidth="md">
      <Card>
        <CardHeader 
          title="Multi-Factor Authentication" 
          subheader="Enhance your account security by enabling MFA"
        />
        <Divider />
        <CardContent>
          {error && (
            <Typography color="error" sx={{ mb: 2 }}>
              {error}
            </Typography>
          )}

          {setupComplete ? (
            renderSetupComplete()
          ) : (
            <>
              <Stepper activeStep={activeStep} sx={{ mb: 4 }}>
                {steps.map((label) => (
                  <Step key={label}>
                    <StepLabel>{label}</StepLabel>
                  </Step>
                ))}
              </Stepper>

              {activeStep === 0 && renderMethodSelection()}
              {activeStep === 1 && renderSetup()}
              {activeStep === 2 && renderVerification()}

              <Box sx={{ display: 'flex', justifyContent: 'space-between', mt: 4 }}>
                <Button
                  disabled={activeStep === 0 || loading}
                  onClick={handleBack}
                >
                  Back
                </Button>
                <Button
                  variant="contained"
                  color="primary"
                  onClick={handleNext}
                  disabled={loading}
                >
                  {loading ? (
                    <CircularProgress size={24} />
                  ) : activeStep === steps.length - 1 ? (
                    'Complete'
                  ) : (
                    'Next'
                  )}
                </Button>
              </Box>
            </>
          )}
        </CardContent>
      </Card>

      {/* Dialog for showing backup codes again */}
      <Dialog
        open={showBackupCodes}
        onClose={() => setShowBackupCodes(false)}
        aria-labelledby="backup-codes-dialog-title"
      >
        <DialogTitle id="backup-codes-dialog-title">
          Your Backup Codes
        </DialogTitle>
        <DialogContent>
          <DialogContentText>
            Save these backup codes in a secure place. Each code can only be used once.
          </DialogContentText>
          <Box sx={{ 
            border: '1px solid #ccc', 
            borderRadius: 1, 
            p: 2, 
            my: 2,
            maxHeight: '200px',
            overflowY: 'auto',
            backgroundColor: '#f5f5f5'
          }}>
            {backupCodes.length > 0 ? (
              <Grid container spacing={2}>
                {backupCodes.map((code, index) => (
                  <Grid item xs={6} key={index}>
                    <Typography variant="body2" fontFamily="monospace">
                      {code}
                    </Typography>
                  </Grid>
                ))}
              </Grid>
            ) : (
              <Typography>No backup codes available</Typography>
            )}
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowBackupCodes(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};

export default MFASetup;

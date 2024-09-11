import {useState} from 'react';
import { SetSecret } from './SetSecret';
import { SetAdminPassword } from './SetAdminPassword';
import React from 'react';

type Props = {
  onCompleted: (newType: boolean) => void;
  cloudType: string;
};

export function SetupBanner({onCompleted, cloudType}:Props) {
  const [step, setStep] = useState<number>(0);
  const [secrets, setSecrets] = useState<SetupResponse>({secret: "", tagHash: "", instanceID: ""});

  React.useEffect(() => {
    if(step === 2) {
      onCompleted(true)
    }
  }, [step]);

  if(step === 0) {
    return <SetSecret onChangeStep={setStep} onChangeSecrets={setSecrets} cloudType={cloudType} />
  } else if(step === 1) {
    return <SetAdminPassword onChangeStep={setStep} secrets={secrets} />
  }
}
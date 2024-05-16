import {useState} from 'react';
import { SetSecret } from './SetSecret';
import { SetAdminPassword } from './SetAdminPassword';
import React from 'react';

type Props = {
  onCompleted: (newType: boolean) => void;
};

export function SetupBanner({onCompleted}:Props) {
  const [step, setStep] = useState<number>(0);
  const [secret, setSecret] = useState<string>("");

  React.useEffect(() => {
    if(step === 2) {
      onCompleted(true)
    }
  }, [step]);

  if(step === 0) {
    return <SetSecret onChangeStep={setStep} onChangeSecret={setSecret} />
  } else if(step === 1) {
    return <SetAdminPassword onChangeStep={setStep} secret={secret} />
  }
}
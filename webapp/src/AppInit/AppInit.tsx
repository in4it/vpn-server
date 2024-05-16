import {
    useQuery,
  } from '@tanstack/react-query'

import React, { useState } from 'react';
import { SetupBanner } from './SetupBanner';
import { AppSettings } from '../Constants/Constants';

  type Props = {
    children?: React.ReactNode
  };

 export const AppInit: React.FC<Props> = ({children}) => {
    const [setupCompleted, setSetupCompleted] = useState<boolean>(false);
    const { isPending, error, data } = useQuery({
      queryKey: ['context'],
      queryFn: () =>
        fetch(AppSettings.url + '/context').then((res) =>
          res.json(),
        ),
    })
    if (isPending) return ''
    if (error) return 'An backend error has occurred: ' + error.message

    if(!setupCompleted && data.setupCompleted) {
      setSetupCompleted(true)
    }

    if (!setupCompleted) {
        return <SetupBanner onCompleted={setSetupCompleted} />
    } else {
        return children
    }
 }
import { NativeSelect, rem, TextInput } from '@mantine/core';
import { useState } from 'react';

type FactorResponse = {
    name: string,
    code: string,
}

type Data = {
    value: string,
    label: string,
}

type Props = {
    setFactorResponse: (newType: FactorResponse) => void,
    captureEnter: (e: React.KeyboardEvent<HTMLDivElement>) => void,
    factors: Array<string>,
};
  

export function MFAInput({setFactorResponse, captureEnter, factors} :Props) {
    const [factorSelected, setSelectedFactor] = useState<string>(factors.length == 0 ? "" : factors[0]);
    const [code, setCode] = useState<string>("");
    const getData = () => {
        const data: Array<Data> = []
        factors.forEach( (element) => {
            data.push({value: element, label: element});
        });
        return data
    }
    const updateFactorResponse = (selectedFactor:string) => {
        setSelectedFactor(selectedFactor)
        setFactorResponse({name: selectedFactor, code: code})
    }
    const updateFactorResponseCode = (code:string) => {
        setCode(code)
        setFactorResponse({name: factorSelected, code: code})
    }
    const select = (
        <NativeSelect
          data={getData()}
          rightSectionWidth={28}
          onChange={(event) => updateFactorResponse(event.currentTarget.value)}
          styles={{
            input: {
              fontWeight: 500,
              borderTopLeftRadius: 0,
              borderBottomLeftRadius: 0,
              width: rem(150),
              marginRight: rem(-2),
            },
          }}
        />
      );
    
      return (
        <TextInput
          type="number"
          placeholder=""
          label="Code"
          onChange={(event) => updateFactorResponseCode(event.currentTarget.value)}
          onKeyDown={(e) => captureEnter(e)}
          rightSection={select}
          rightSectionWidth={150}
        />
      );
    
}
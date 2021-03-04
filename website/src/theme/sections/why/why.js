import React from 'react';
import useBaseUrl from '@docusaurus/useBaseUrl';
import Heading from './../../components/heading/heading';
import Container from './../../components/container/container';

export default function Why() {
  const component = 'shift-why';

  return (
    <div className={component}>
      <Container
        componentClass={component}
        size={'small'}
      >
        <Heading
          componentClass={component}
          title={'Why use OpenState?'}
          subtitle={'OpenState is adopted from and designed for production environments seeking flexible, low-code workflow management. Declarative task definitions mean you can spend more time elsewhere, while OpenState makes light work of your ETL jobs, automation, and closed loop control planes.'}
          align={'left'}
          titleSize={'medium'}
        />
      </Container>
      <Container
        componentClass={component}
        size={'medium'}
      >
      </Container>
    </div>
  )
}

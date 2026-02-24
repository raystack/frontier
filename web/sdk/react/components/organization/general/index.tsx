'use client';

import { GeneralPage } from '~/react/views/general';

export default function GeneralSetting() {
  return <GeneralPage onDeleteSuccess={() => {
    // @ts-ignore
    window.location = window.location.origin;
  }}/>;
}

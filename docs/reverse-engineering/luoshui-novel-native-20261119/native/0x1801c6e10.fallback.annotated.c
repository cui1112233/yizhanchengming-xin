/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1801c6e10 */

longlong FUN_1801c6e10(longlong param_1,longlong param_2,longlong *param_3)

{
  undefined8 *puVar1;
  int *piVar2;
  code *pcVar3;
  undefined8 uVar4;
  char cVar5;
  int iVar6;
  longlong *plVar7;
  longlong lVar8;
  longlong *plVar9;
  undefined8 local_res10 [2];
  longlong local_res20;
  
  plVar7 = DAT_1803bbd48;
  uVar4 = DAT_180391dc8;
  puVar1 = *(undefined8 **)(param_2 + 0x90);
  if (*(int *)(param_2 + 0x78) == 1) {
    FUN_180371770((longlong)puVar1 + 0x24,&local_res20,8,local_res10,8,0);
    if (param_3 != (longlong *)0x0) {
      if ((-1 < (int)*param_3) && (*param_3 = *param_3 + -1, *param_3 == 0)) {
        (**(code **)(param_3[1] + 0x30))(param_3);
      }
      cVar5 = FUN_18000d5b0(param_1);
      if (cVar5 != '\0') goto LAB_1801c6eeb;
    }
LAB_1801c70ee:
    puVar1[3] = *(undefined8 *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    *puVar1 = 0;
    puVar1[1] = 0;
    puVar1[2] = 0;
    puVar1[3] = uVar4;
    *(undefined4 *)(puVar1 + 4) = 0;
    if (plVar7 == (longlong *)0x0) {
LAB_1801c6e86:
      plVar7 = (longlong *)FUN_18036cd10(param_1,DAT_1803bbdb0,DAT_1803ca610,0x18);
      DAT_1803bbd48 = plVar7;
    }
    else if ((1 < *plVar7) || (plVar7[2] != 0)) {
      if ((plVar7 != (longlong *)0x0) &&
         ((-1 < (int)*plVar7 && (*plVar7 = *plVar7 + -1, *plVar7 == 0)))) {
        (**(code **)(plVar7[1] + 0x30))();
      }
      goto LAB_1801c6e86;
    }
    *(longlong **)(param_2 + 0x50) = plVar7;
    iVar6 = (int)*plVar7 + 1;
    if (iVar6 != 0) {
      *(int *)plVar7 = iVar6;
    }
    lVar8 = *(longlong *)(param_2 + 0x50);
    *(longlong *)(lVar8 + 0x38) = param_2;
    *(undefined4 *)(lVar8 + 0x40) = 0;
    FUN_180001df0(param_1,*(undefined8 *)(param_2 + 0x50));
    piVar2 = *(int **)(param_1 + 0xf8);
    *(int **)(param_2 + 0x68) = piVar2;
    if ((piVar2 != (int *)0x0) && (*piVar2 + 1 != 0)) {
      *piVar2 = *piVar2 + 1;
    }
LAB_1801c6eeb:
    lVar8 = *(longlong *)(*(longlong *)(*(longlong *)(param_2 + 0xa0) + 0x10) + 8);
    pcVar3 = *(code **)(lVar8 + 0xe0);
    if (pcVar3 == (code *)0x0) {
      PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,"\'%s\' object is not an iterator",
                   *(undefined8 *)(lVar8 + 0x18));
LAB_1801c6f9d:
      cVar5 = FUN_18000cd20(param_1,*(undefined8 *)PyExc_StopIteration_exref);
      if (cVar5 != '\0') {
        plVar7 = *(longlong **)(param_2 + 0x68);
        if (((plVar7 != (longlong *)0x0) && (*(undefined8 *)(param_2 + 0x68) = 0, -1 < (int)*plVar7)
            ) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))();
        }
        plVar7 = (longlong *)puVar1[1];
        if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))();
        }
        plVar7 = (longlong *)*puVar1;
        puVar1[1] = 0;
        if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))();
        }
        *puVar1 = 0;
        return 0;
      }
      goto LAB_1801c70ee;
    }
    lVar8 = (*pcVar3)();
    if (lVar8 == 0) goto LAB_1801c6f9d;
    plVar7 = (longlong *)puVar1[1];
    puVar1[1] = lVar8;
    if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    piVar2 = (int *)puVar1[1];
    plVar7 = (longlong *)*puVar1;
    *puVar1 = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    lVar8 = *(longlong *)(*(longlong *)(param_2 + 0xa8) + 0x10);
    if (lVar8 != 0) {
      local_res10[0] = *puVar1;
      *(undefined4 *)(*(longlong *)(param_2 + 0x50) + 0x28) = 0x81;
      local_res20 = lVar8;
      lVar8 = FUN_180376770(param_1,lVar8,local_res10[0]);
      if (lVar8 != 0) {
        FUN_180371700((longlong)puVar1 + 0x24,&local_res20,8,local_res10,8,0);
        *(undefined4 *)(param_2 + 0x78) = 1;
        return lVar8;
      }
      goto LAB_1801c70ee;
    }
/* novels.prompts.inkos.reviser_prompts#29="is_local" */
    FUN_180371050(param_1,puVar1 + 3,DAT_1803bbef8);
    FUN_18000d1d0(param_1,puVar1[3]);
  }
  plVar9 = puVar1 + 3;
  *(undefined4 *)(puVar1 + 4) = 0x81;
  puVar1[2] = &DAT_180394c38;
  plVar7 = (longlong *)puVar1[1];
  *(undefined4 *)(puVar1 + 0x86) = 0x81;
  *(undefined4 *)(puVar1 + 4) = 0;
  puVar1[0x85] = *plVar9;
  *plVar9 = 0;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  *plVar9 = puVar1[0x85];
  *(undefined4 *)(puVar1 + 4) = *(undefined4 *)(puVar1 + 0x86);
  puVar1[1] = 0;
  cVar5 = FUN_180001380(param_1,plVar9);
  if (cVar5 != '\0') goto LAB_1801c71f8;
  piVar2 = *(int **)(*plVar9 + 0x28);
  if (piVar2 == (int *)0x0) {
    lVar8 = FUN_18038e2a0(*(undefined8 *)(param_2 + 0x50));
LAB_1801c71a5:
    FUN_180001300(plVar9,lVar8);
  }
  else if ((*(int *)(puVar1 + 4) != 0) &&
          (*(longlong *)(piVar2 + 6) != *(longlong *)(param_2 + 0x50))) {
    lVar8 = FUN_18038e2a0();
    *(int **)(lVar8 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    goto LAB_1801c71a5;
  }
  FUN_18036d0a0(*(undefined8 *)(param_2 + 0x50),puVar1[2],0,*puVar1,*(undefined8 *)(param_2 + 0xa8))
  ;
  plVar7 = DAT_1803bbd48;
  if (*(longlong **)(param_2 + 0x50) == DAT_1803bbd48) {
    if ((-1 < (int)*DAT_1803bbd48) && (*DAT_1803bbd48 = *DAT_1803bbd48 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    DAT_1803bbd48 = (longlong *)0x0;
  }
LAB_1801c71f8:
  plVar7 = *(longlong **)(param_2 + 0x68);
  if (((plVar7 != (longlong *)0x0) && (*(undefined8 *)(param_2 + 0x68) = 0, -1 < (int)*plVar7)) &&
     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  plVar7 = (longlong *)*puVar1;
  *(undefined4 *)(puVar1 + 0x88) = *(undefined4 *)(puVar1 + 4);
  puVar1[0x87] = *plVar9;
  *(undefined4 *)(puVar1 + 4) = 0;
  *plVar9 = 0;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  *plVar9 = puVar1[0x87];
  *puVar1 = 0;
  *(undefined4 *)(puVar1 + 4) = *(undefined4 *)(puVar1 + 0x88);
  plVar7 = *(longlong **)(param_1 + 0x60);
  *(undefined8 *)(param_1 + 0x60) = puVar1[0x87];
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  return 0;
}

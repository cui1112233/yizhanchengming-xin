/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180052020 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

longlong FUN_180052020(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  int *piVar1;
  undefined8 uVar2;
  undefined8 uVar3;
  int iVar4;
  longlong lVar5;
  longlong *plVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong *plVar11;
  longlong *plVar12;
  longlong lVar13;
  code *pcVar14;
  undefined8 uVar15;
  longlong *plVar16;
  undefined1 auStack_378 [32];
  longlong *local_358;
  longlong *local_350;
  longlong *local_348;
  longlong *local_340;
  longlong *local_338;
  longlong *local_330;
  longlong *local_328;
  longlong *local_320;
  longlong *local_318;
  longlong *local_310;
  longlong *local_308;
  longlong *local_300;
  longlong *local_2f8;
  longlong *local_2f0;
  longlong *local_2e8;
  longlong *local_2e0;
  longlong *local_2d8;
  longlong *local_2d0;
  longlong *local_2c8;
  longlong *local_2c0;
  longlong *local_2b8;
  longlong local_2b0;
  longlong *local_2a8;
  longlong *local_2a0;
  longlong *local_298;
  longlong *local_290;
  longlong *local_280;
  longlong local_278;
  longlong *local_270;
  longlong *local_268;
  longlong *local_260;
  longlong *local_258;
  longlong *local_250;
  longlong *local_248;
  longlong *local_240;
  longlong *local_238;
  longlong *local_230;
  longlong *local_228;
  longlong *local_220;
  longlong *local_218;
  longlong *local_210;
  longlong *local_208;
  longlong *local_200;
  longlong *local_1f8;
  longlong *local_1f0;
  longlong *local_1e8;
  longlong *local_1e0;
  longlong *local_1d8;
  longlong *local_1d0;
  longlong *local_1c8;
  longlong *local_1c0;
  longlong *local_1b8;
  longlong *local_1b0;
  longlong *local_1a8;
  longlong *local_1a0;
  longlong *local_198;
  longlong *local_190;
  longlong *local_188;
  longlong *local_180;
  longlong *local_178;
  longlong *local_170;
  longlong *local_168;
  longlong *local_160;
  longlong *local_158;
  longlong *local_150;
  longlong *local_148;
  longlong *local_140;
  longlong *local_138;
  longlong *local_130;
  longlong *local_128;
  longlong *local_120;
  longlong *local_118;
  longlong *local_110;
  longlong *local_108;
  longlong *local_100;
  longlong *local_f8;
  longlong *local_f0;
  longlong *local_e8;
  longlong *local_e0;
  longlong *local_d8;
  longlong *local_d0;
  longlong *local_c8;
  longlong *local_c0;
  longlong *local_b8;
  longlong *local_b0;
  longlong *local_a8;
  longlong *local_a0;
  longlong *local_98;
  longlong *local_90;
  longlong *local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  ulonglong local_48;
  
  plVar8 = DAT_1803aec80;
  local_48 = DAT_1803a6040 ^ (ulonglong)auStack_378;
  local_218 = (longlong *)*param_3;
  plVar10 = (longlong *)param_3[6];
  plVar6 = (longlong *)param_3[2];
  local_210 = (longlong *)param_3[1];
  local_238 = (longlong *)param_3[3];
  local_208 = (longlong *)param_3[4];
  local_200 = (longlong *)param_3[5];
  local_1c8 = (longlong *)param_3[7];
  plVar7 = (longlong *)0x0;
  plVar16 = (longlong *)0x0;
  local_2b0 = DAT_180391c98;
  local_1b0 = (longlong *)0x0;
  local_1a8 = (longlong *)0x0;
  local_1f8 = (longlong *)0x0;
  local_1f0 = (longlong *)0x0;
  local_1e8 = (longlong *)0x0;
  local_1e0 = (longlong *)0x0;
  local_248 = (longlong *)0x0;
  local_1d8 = (longlong *)0x0;
  local_230 = (longlong *)0x0;
  local_228 = (longlong *)0x0;
  local_240 = (longlong *)0x0;
  local_1d0 = (longlong *)0x0;
  local_220 = (longlong *)0x0;
  local_2a0 = DAT_1803aec80;
  local_2a8 = plVar6;
  local_298 = plVar10;
  local_278 = param_1;
  if (DAT_1803aec80 == (longlong *)0x0) {
LAB_180052117:
    DAT_1803aec80 = (longlong *)FUN_18036cd10(param_1,DAT_1803aec90,DAT_1803ca868,0xb0);
  }
  else if ((1 < *DAT_1803aec80) || (DAT_1803aec80[2] != 0)) {
    if ((DAT_1803aec80 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803aec80 && (*DAT_1803aec80 = *DAT_1803aec80 + -1, *plVar8 == 0)))) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    goto LAB_180052117;
  }
  plVar8 = DAT_1803aec80;
  local_2a0 = DAT_1803aec80;
  FUN_180001d00(param_1,DAT_1803aec80);
  lVar5 = FUN_180051970(param_1);
  if (lVar5 == 0) {
    FUN_180373340(param_1,&local_2b0,DAT_1803aed28);
    uVar15 = 0x2f;
    plVar16 = plVar7;
  }
  else {
    uVar15 = 0x2f;
    *(undefined4 *)(plVar8 + 5) = 0x2f;
    local_1c0 = (longlong *)FUN_180376770(param_1,lVar5,plVar6);
    if (local_1c0 == (longlong *)0x0) {
      local_2b0 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      local_1b0 = local_1c0;
      lVar5 = FUN_180051a80(param_1);
      if (lVar5 == 0) {
        FUN_180373340(param_1,&local_2b0,DAT_1803aed30);
        uVar15 = 0x30;
        plVar8 = local_2a0;
        plVar16 = plVar7;
      }
      else {
        plVar6 = (longlong *)FUN_18037f0c0(param_1,local_210,DAT_1803aed38);
        uVar15 = 0x30;
        if (plVar6 == (longlong *)0x0) {
          local_2b0 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          plVar8 = local_2a0;
          plVar16 = plVar7;
          plVar6 = local_2a8;
        }
        else {
          *(undefined4 *)(local_2a0 + 5) = 0x30;
          plVar7 = (longlong *)FUN_180376770(param_1,lVar5,plVar6);
          local_1a0 = plVar7;
          if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
            (**(code **)(plVar6[1] + 0x30))(plVar6);
          }
          if (plVar7 == (longlong *)0x0) {
            local_2b0 = *(longlong *)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar8 = local_2a0;
            plVar6 = local_2a8;
          }
          else {
            local_1a8 = plVar7;
            lVar5 = FUN_180051420(param_1);
            plVar6 = local_210;
            plVar8 = local_2a0;
            if (lVar5 == 0) {
              FUN_180373340(param_1,&local_2b0,DAT_1803aed40);
              uVar15 = 0x31;
              plVar8 = local_2a0;
              plVar6 = local_2a8;
            }
            else {
              uVar15 = 0x31;
              *(undefined4 *)(local_2a0 + 5) = 0x31;
              local_198 = (longlong *)FUN_180376770(param_1,lVar5,local_210);
              if (local_198 == (longlong *)0x0) {
                local_2b0 = *(longlong *)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar6 = local_2a8;
              }
              else {
                local_1f8 = local_198;
                plVar6 = (longlong *)FUN_18037f0c0(param_1,plVar6,DAT_1803aed48);
                if (plVar6 == (longlong *)0x0) {
                  local_2b0 = *(longlong *)(param_1 + 0x60);
                  uVar15 = 0x32;
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  plVar6 = local_2a8;
                }
                else {
                  iVar4 = FUN_180001510(plVar6);
                  if (iVar4 == -1) {
                    local_2b0 = *(longlong *)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                      (**(code **)(plVar6[1] + 0x30))(plVar6);
                    }
                    uVar15 = 0x32;
                    plVar6 = local_2a8;
                  }
                  else {
                    if (iVar4 != 1) {
                      if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                        (**(code **)(plVar6[1] + 0x30))(plVar6);
                      }
                      plVar6 = DAT_1803aed50;
                      iVar4 = (int)*DAT_1803aed50 + 1;
                      if (iVar4 != 0) {
                        *(int *)DAT_1803aed50 = iVar4;
                      }
                    }
                    iVar4 = FUN_1803836d0(plVar6,DAT_1803aed58);
                    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                      (**(code **)(plVar6[1] + 0x30))(plVar6);
                    }
                    if (iVar4 == -1) {
                      local_2b0 = *(longlong *)(param_1 + 0x60);
                      uVar15 = 0x32;
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      plVar6 = local_2a8;
                    }
                    else {
                      plVar16 = DAT_1803aed50;
                      if (iVar4 == 1) {
                        plVar16 = DAT_1803aed58;
                      }
                      iVar4 = (int)*plVar16 + 1;
                      if (iVar4 != 0) {
                        *(int *)plVar16 = iVar4;
                      }
                      local_290 = plVar16;
                      plVar6 = (longlong *)FUN_18037f0c0(param_1,plVar10,DAT_1803aed60);
                      uVar15 = 0x34;
                      if (plVar6 == (longlong *)0x0) {
LAB_180052413:
                        local_2b0 = *(longlong *)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        plVar6 = local_2a8;
                      }
                      else {
                        *(undefined4 *)(plVar8 + 5) = 0x34;
                        plVar8 = (longlong *)FUN_180376210(param_1,plVar6);
                        if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                          (**(code **)(plVar6[1] + 0x30))(plVar6);
                        }
                        if (plVar8 != (longlong *)0x0) {
                          iVar4 = FUN_180001510(plVar8);
                          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                            (**(code **)(plVar8[1] + 0x30))(plVar8);
                          }
                          plVar6 = local_218;
                          plVar8 = local_2a0;
                          if (iVar4 != -1) {
                            plVar9 = plVar10;
                            if (iVar4 == 0) {
                              local_118 = local_210;
                              uVar15 = 0x35;
                              local_100 = local_238;
                              local_f8 = local_208;
                              *(undefined4 *)(local_2a0 + 5) = 0x35;
                              local_f0 = local_200;
                              local_110 = plVar7;
                              local_108 = plVar16;
                              plVar9 = (longlong *)
                                       FUN_18037eda0(param_1,local_218,DAT_1803aed68,&local_118);
                              if (plVar9 == (longlong *)0x0) goto LAB_180052413;
                              local_298 = plVar9;
                              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0))
                              {
                                (**(code **)(plVar10[1] + 0x30))(plVar10);
                              }
                            }
                            plVar10 = (longlong *)FUN_18037f0c0(param_1,plVar6,DAT_1803aed20);
                            if (plVar10 != (longlong *)0x0) {
                              plVar7 = (longlong *)FUN_18037f0c0(param_1,plVar10,DAT_1803aed70);
                              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0))
                              {
                                (**(code **)(plVar10[1] + 0x30))(plVar10);
                              }
                              piVar1 = DAT_1803aed88;
                              plVar6 = DAT_1803aed80;
                              plVar10 = DAT_1803aed78;
                              if (plVar7 != (longlong *)0x0) {
                                plVar9 = (longlong *)FUN_180374a40(param_1,3);
                                if (*piVar1 + 1 != 0) {
                                  *piVar1 = *piVar1 + 1;
                                }
                                plVar9[3] = (longlong)piVar1;
                                plVar8 = DAT_1803aed90;
                                if (DAT_1803ab740 == 0) {
                                  DAT_1803ab740 = FUN_18035f440("format");
                                }
                                local_270 = local_238;
                                local_268 = plVar8;
                                lVar5 = FUN_180377750(param_1,DAT_1803ab740,&local_270);
                                if (lVar5 == 0) {
                                  uVar15 = 0x3b;
LAB_180052605:
                                  local_2b0 = *(longlong *)(param_1 + 0x60);
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0))
                                  {
                                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                                  }
                                  plVar8 = local_2a0;
                                  plVar16 = local_290;
                                  plVar6 = local_2a8;
                                  plVar10 = local_298;
                                  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                  {
                                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                                    plVar8 = local_2a0;
                                    plVar16 = local_290;
                                    plVar6 = local_2a8;
                                    plVar10 = local_298;
                                  }
                                }
                                else {
                                  plVar9[4] = lVar5;
                                  piVar1 = DAT_1803aed98;
                                  if (*DAT_1803aed98 + 1 != 0) {
                                    *DAT_1803aed98 = *DAT_1803aed98 + 1;
                                  }
                                  plVar9[5] = (longlong)piVar1;
                                  plVar8 = (longlong *)PyUnicode_Join(DAT_1803aed90,plVar9);
                                  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                  {
                                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                                  }
                                  if (plVar8 == (longlong *)0x0) {
                                    local_2b0 = *(longlong *)(param_1 + 0x60);
                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                    if ((-1 < (int)*plVar7) &&
                                       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                      (**(code **)(plVar7[1] + 0x30))(plVar7);
                                    }
                                    uVar15 = 0x3b;
                                    plVar8 = local_2a0;
                                    plVar16 = local_290;
                                    plVar6 = local_2a8;
                                    plVar10 = local_298;
                                  }
                                  else {
                                    uVar15 = 0x37;
                                    local_150 = plVar10;
                                    local_148 = plVar6;
                                    *(undefined4 *)(local_2a0 + 5) = 0x37;
                                    local_358 = (longlong *)DAT_1803aeda8;
                                    local_140 = plVar8;
                                    plVar10 = (longlong *)
                                              FUN_18037d280(param_1,plVar7,DAT_1803aeda0,&local_150)
                                    ;
                                    if ((-1 < (int)*plVar7) &&
                                       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                      (**(code **)(plVar7[1] + 0x30))(plVar7);
                                    }
                                    if ((-1 < (int)*plVar8) &&
                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                      (**(code **)(plVar8[1] + 0x30))(plVar8);
                                    }
                                    if (plVar10 == (longlong *)0x0) {
                                      local_2b0 = *(longlong *)(param_1 + 0x60);
                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                      plVar8 = local_2a0;
                                      plVar16 = local_290;
                                      plVar6 = local_2a8;
                                      plVar10 = local_298;
                                    }
                                    else {
                                      if ((-1 < (int)*plVar10) &&
                                         (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                        (**(code **)(plVar10[1] + 0x30))(plVar10);
                                      }
                                      local_270 = (longlong *)FUN_180050f40(param_1);
                                      if (local_270 == (longlong *)0x0) {
                                        FUN_180373340(param_1,&local_2b0,DAT_1803aedb0);
                                        uVar15 = 0x3e;
                                        plVar8 = local_2a0;
                                        plVar16 = local_290;
                                        plVar6 = local_2a8;
                                        plVar10 = local_298;
                                      }
                                      else {
                                        plVar7 = (longlong *)
                                                 FUN_180001b80(param_1,local_200,DAT_1803aedb8);
                                        if (plVar7 == (longlong *)0x0) {
                                          local_2b0 = *(longlong *)(param_1 + 0x60);
                                          uVar15 = 0x41;
                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                          plVar8 = local_2a0;
                                          plVar16 = local_290;
                                          plVar6 = local_2a8;
                                          plVar10 = local_298;
                                        }
                                        else {
                                          plVar9 = (longlong *)
                                                   FUN_18037f0c0(param_1,local_1c0,DAT_1803aedc0);
                                          if (plVar9 == (longlong *)0x0) {
                                            local_2b0 = *(longlong *)(param_1 + 0x60);
                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                            if ((-1 < (int)*plVar7) &&
                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                              pcVar14 = *(code **)(plVar7[1] + 0x30);
                                              plVar9 = plVar7;
LAB_180052802:
                                              (*pcVar14)(plVar9);
                                            }
                                          }
                                          else {
                                            iVar4 = FUN_180001510(plVar9);
                                            if (iVar4 != -1) {
                                              if (iVar4 != 1) {
                                                if ((-1 < (int)*plVar9) &&
                                                   (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                  (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                }
                                                plVar9 = DAT_1803aedc8;
                                                iVar4 = (int)*DAT_1803aedc8 + 1;
                                                if (iVar4 != 0) {
                                                  *(int *)DAT_1803aedc8 = iVar4;
                                                }
                                              }
                                              lVar5 = FUN_180050e30(param_1);
                                              if (lVar5 == 0) {
                                                FUN_180373340(param_1,&local_2b0,DAT_1803aedd0);
                                              }
                                              else {
                                                plVar10 = (longlong *)
                                                          FUN_18037f0c0(param_1,lVar5,DAT_1803aedd8)
                                                ;
                                                plVar6 = local_2a8;
                                                if (plVar10 != (longlong *)0x0) {
                                                  plVar8 = (longlong *)
                                                           FUN_180382440(local_2a8,DAT_1803aede0);
                                                  if (plVar8 == (longlong *)0x0) {
                                                    local_2b0 = *(longlong *)(param_1 + 0x60);
                                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                                    if ((-1 < (int)*plVar7) &&
                                                       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                      (**(code **)(plVar7[1] + 0x30))(plVar7);
                                                    }
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                    }
                                                    if ((-1 < (int)*plVar10) &&
                                                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                      (**(code **)(plVar10[1] + 0x30))(plVar10);
                                                    }
                                                    uVar15 = 0x43;
                                                    plVar8 = local_2a0;
                                                    plVar16 = local_290;
                                                    plVar10 = local_298;
                                                    goto LAB_180052819;
                                                  }
                                                  uVar15 = 0x43;
                                                  local_188 = DAT_1803aed90;
                                                  *(undefined4 *)(local_2a0 + 5) = 0x43;
                                                  local_190 = plVar8;
                                                  plVar6 = (longlong *)
                                                           FUN_180377750(param_1,plVar10,&local_190)
                                                  ;
                                                  if ((-1 < (int)*plVar10) &&
                                                     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                    (**(code **)(plVar10[1] + 0x30))(plVar10);
                                                  }
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                  }
                                                  plVar10 = local_1c0;
                                                  if (plVar6 == (longlong *)0x0) goto LAB_180052605;
                                                  plVar8 = (longlong *)
                                                           FUN_18037f0c0(param_1,local_1c0,
                                                                         DAT_1803aede8);
                                                  if (plVar8 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_2b0);
                                                    if ((-1 < (int)*plVar7) &&
                                                       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                      (**(code **)(plVar7[1] + 0x30))(plVar7);
                                                    }
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                    }
                                                    if ((-1 < (int)*plVar6) &&
                                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                      (**(code **)(plVar6[1] + 0x30))(plVar6);
                                                    }
                                                  }
                                                  else {
                                                    iVar4 = FUN_180001510(plVar8);
                                                    if (iVar4 != -1) {
                                                      if (iVar4 != 1) {
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar5 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar5 + 0x30))(plVar8);
                                                        }
                                                        plVar8 = DAT_1803aedc8;
                                                        iVar4 = (int)*DAT_1803aedc8 + 1;
                                                        if (iVar4 != 0) {
                                                          *(int *)DAT_1803aedc8 = iVar4;
                                                        }
                                                      }
                                                      plVar11 = (longlong *)
                                                                FUN_18037f0c0(param_1,plVar10,
                                                                              DAT_1803aedf0);
                                                      if (plVar11 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_2b0);
                                                        if ((-1 < (int)*plVar7) &&
                                                           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                          lVar5 = FUN_18000cc40(plVar7);
                                                          (**(code **)(lVar5 + 0x30))(plVar7);
                                                        }
                                                        if ((-1 < (int)*plVar9) &&
                                                           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                          lVar5 = FUN_18000cc40(plVar9);
                                                          (**(code **)(lVar5 + 0x30))(plVar9);
                                                        }
                                                        if ((-1 < (int)*plVar6) &&
                                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                          lVar5 = FUN_18000cc40(plVar6);
                                                          (**(code **)(lVar5 + 0x30))(plVar6);
                                                        }
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar5 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar5 + 0x30))(plVar8);
                                                        }
                                                      }
                                                      else {
                                                        iVar4 = FUN_180001510(plVar11);
                                                        if (iVar4 != -1) {
                                                          if (iVar4 != 1) {
                                                            if ((-1 < (int)*plVar11) &&
                                                               (*plVar11 = *plVar11 + -1,
                                                               *plVar11 == 0)) {
                                                              lVar5 = FUN_18000cc40(plVar11);
                                                              (**(code **)(lVar5 + 0x30))(plVar11);
                                                            }
                                                            plVar11 = DAT_1803aedf8;
                                                            iVar4 = (int)*DAT_1803aedf8 + 1;
                                                            if (iVar4 != 0) {
                                                              *(int *)DAT_1803aedf8 = iVar4;
                                                            }
                                                          }
                                                          lVar5 = FUN_180050e30(param_1);
                                                          if (lVar5 == 0) {
                                                            FUN_180373340(param_1,&local_2b0,
                                                                          DAT_1803aedd0);
LAB_180052d3f:
                                                            if ((-1 < (int)*plVar7) &&
                                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0
                                                               )) {
                                                              lVar5 = FUN_18000cc40(plVar7);
                                                              (**(code **)(lVar5 + 0x30))(plVar7);
                                                            }
                                                            if ((-1 < (int)*plVar9) &&
                                                               (*plVar9 = *plVar9 + -1, *plVar9 == 0
                                                               )) {
                                                              lVar5 = FUN_18000cc40(plVar9);
                                                              (**(code **)(lVar5 + 0x30))(plVar9);
                                                            }
                                                            if ((-1 < (int)*plVar6) &&
                                                               (*plVar6 = *plVar6 + -1, *plVar6 == 0
                                                               )) {
                                                              lVar5 = FUN_18000cc40(plVar6);
                                                              (**(code **)(lVar5 + 0x30))(plVar6);
                                                            }
                                                            if ((-1 < (int)*plVar8) &&
                                                               (*plVar8 = *plVar8 + -1, *plVar8 == 0
                                                               )) {
                                                              lVar5 = FUN_18000cc40(plVar8);
                                                              (**(code **)(lVar5 + 0x30))(plVar8);
                                                            }
                                                            if ((-1 < (int)*plVar11) &&
                                                               (*plVar11 = *plVar11 + -1,
                                                               *plVar11 == 0)) {
                                                              lVar5 = FUN_18000cc40(plVar11);
                                                              (**(code **)(lVar5 + 0x30))(plVar11);
                                                            }
                                                          }
                                                          else {
                                                            plVar10 = (longlong *)
                                                                      FUN_18037f0c0(param_1,lVar5,
                                                                                    DAT_1803aedd8);
                                                            local_250 = plVar10;
                                                            if (plVar10 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_2b0);
                                                              goto LAB_180052d3f;
                                                            }
                                                            local_280 = (longlong *)
                                                                        FUN_180382540(local_2a8,
                                                                                      DAT_1803aee00)
                                                            ;
                                                            if (local_280 != (longlong *)0x0) {
                                                              uVar15 = 0x46;
                                                              local_178 = DAT_1803aedf8;
                                                              *(undefined4 *)(local_2a0 + 5) = 0x46;
                                                              local_180 = local_280;
                                                              plVar12 = (longlong *)
                                                                        FUN_180377750(param_1,
                                                  local_250,&local_180);
                                                  if ((-1 < (int)*local_250) &&
                                                     (*local_250 = *local_250 + -1, *local_250 == 0)
                                                     ) {
                                                    lVar5 = FUN_18000cc40(local_250);
                                                    (**(code **)(lVar5 + 0x30))(local_250);
                                                  }
                                                  if ((-1 < (int)*local_280) &&
                                                     (*local_280 = *local_280 + -1, *local_280 == 0)
                                                     ) {
                                                    lVar5 = FUN_18000cc40();
                                                    (**(code **)(lVar5 + 0x30))(local_280);
                                                  }
                                                  param_1 = local_278;
                                                  if (plVar12 == (longlong *)0x0) {
                                                    FUN_18000ce30(local_278,&local_2b0);
                                                    if ((-1 < (int)*plVar7) &&
                                                       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                      lVar5 = FUN_18000cc40(plVar7);
                                                      (**(code **)(lVar5 + 0x30))(plVar7);
                                                    }
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      lVar5 = FUN_18000cc40(plVar9);
                                                      (**(code **)(lVar5 + 0x30))(plVar9);
                                                    }
                                                    if ((-1 < (int)*plVar6) &&
                                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                      lVar5 = FUN_18000cc40(plVar6);
                                                      (**(code **)(lVar5 + 0x30))(plVar6);
                                                    }
                                                    if ((-1 < (int)*plVar8) &&
                                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                      lVar5 = FUN_18000cc40(plVar8);
                                                      (**(code **)(lVar5 + 0x30))(plVar8);
                                                    }
                                                    plVar8 = local_2a0;
                                                    plVar16 = local_290;
                                                    plVar6 = local_2a8;
                                                    plVar10 = local_298;
                                                    if ((-1 < (int)*plVar11) &&
                                                       (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                      lVar5 = FUN_18000cc40(plVar11);
                                                      (**(code **)(lVar5 + 0x30))(plVar11);
                                                      plVar8 = local_2a0;
                                                      plVar16 = local_290;
                                                      plVar6 = local_2a8;
                                                      plVar10 = local_298;
                                                    }
                                                    goto LAB_180052819;
                                                  }
                                                  lVar5 = FUN_180050e30(local_278);
                                                  if (lVar5 == 0) {
                                                    FUN_180373340(param_1,&local_2b0,DAT_1803aedd0);
LAB_180053046:
                                                    if ((-1 < (int)*plVar7) &&
                                                       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                      lVar5 = FUN_18000cc40(plVar7);
                                                      (**(code **)(lVar5 + 0x30))(plVar7);
                                                    }
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      lVar5 = FUN_18000cc40(plVar9);
                                                      (**(code **)(lVar5 + 0x30))(plVar9);
                                                    }
                                                    if ((-1 < (int)*plVar6) &&
                                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                      lVar5 = FUN_18000cc40(plVar6);
                                                      (**(code **)(lVar5 + 0x30))(plVar6);
                                                    }
                                                    if ((-1 < (int)*plVar8) &&
                                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                      lVar5 = FUN_18000cc40(plVar8);
                                                      (**(code **)(lVar5 + 0x30))(plVar8);
                                                    }
                                                    if ((-1 < (int)*plVar11) &&
                                                       (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                      lVar5 = FUN_18000cc40(plVar11);
                                                      (**(code **)(lVar5 + 0x30))(plVar11);
                                                    }
                                                    if ((-1 < (int)*plVar12) &&
                                                       (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                      lVar5 = FUN_18000cc40(plVar12);
                                                      (**(code **)(lVar5 + 0x30))(plVar12);
                                                    }
                                                  }
                                                  else {
                                                    local_258 = (longlong *)
                                                                FUN_18037f0c0(param_1,lVar5,
                                                                              DAT_1803aedd8);
                                                    if (local_258 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_2b0);
                                                      goto LAB_180053046;
                                                    }
                                                    local_280 = (longlong *)
                                                                FUN_180382540(local_2a8,
                                                                              DAT_1803aee08);
                                                    if (local_280 != (longlong *)0x0) {
                                                      uVar15 = 0x47;
                                                      local_168 = DAT_1803aedf8;
                                                      *(undefined4 *)(local_2a0 + 5) = 0x47;
                                                      local_170 = local_280;
                                                      local_260 = (longlong *)
                                                                  FUN_180377750(local_278,local_258,
                                                                                &local_170);
                                                      if ((-1 < (int)*local_258) &&
                                                         (*local_258 = *local_258 + -1,
                                                         *local_258 == 0)) {
                                                        lVar5 = FUN_18000cc40(local_258);
                                                        (**(code **)(lVar5 + 0x30))(local_258);
                                                      }
                                                      if ((-1 < (int)*local_280) &&
                                                         (*local_280 = *local_280 + -1,
                                                         *local_280 == 0)) {
                                                        lVar5 = FUN_18000cc40();
                                                        (**(code **)(lVar5 + 0x30))(local_280);
                                                      }
                                                      lVar5 = local_278;
                                                      if (local_260 == (longlong *)0x0) {
                                                        FUN_18000ce30(local_278,&local_2b0);
                                                        if ((-1 < (int)*plVar7) &&
                                                           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                          lVar5 = FUN_18000cc40(plVar7);
                                                          (**(code **)(lVar5 + 0x30))(plVar7);
                                                        }
                                                        if ((-1 < (int)*plVar9) &&
                                                           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                          lVar5 = FUN_18000cc40(plVar9);
                                                          (**(code **)(lVar5 + 0x30))(plVar9);
                                                        }
                                                        if ((-1 < (int)*plVar6) &&
                                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                          lVar5 = FUN_18000cc40(plVar6);
                                                          (**(code **)(lVar5 + 0x30))(plVar6);
                                                        }
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar5 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar5 + 0x30))(plVar8);
                                                        }
                                                        if ((-1 < (int)*plVar11) &&
                                                           (*plVar11 = *plVar11 + -1, *plVar11 == 0)
                                                           ) {
                                                          lVar5 = FUN_18000cc40(plVar11);
                                                          (**(code **)(lVar5 + 0x30))(plVar11);
                                                        }
                                                        plVar8 = local_2a0;
                                                        plVar16 = local_290;
                                                        param_1 = local_278;
                                                        plVar6 = local_2a8;
                                                        plVar10 = local_298;
                                                        if ((-1 < (int)*plVar12) &&
                                                           (*plVar12 = *plVar12 + -1, *plVar12 == 0)
                                                           ) {
                                                          lVar5 = FUN_18000cc40(plVar12);
                                                          (**(code **)(lVar5 + 0x30))(plVar12);
                                                          plVar8 = local_2a0;
                                                          plVar16 = local_290;
                                                          param_1 = local_278;
                                                          plVar6 = local_2a8;
                                                          plVar10 = local_298;
                                                        }
                                                        goto LAB_180052819;
                                                      }
                                                      lVar13 = FUN_180050e30(local_278);
                                                      if (lVar13 == 0) {
                                                        FUN_180373340(lVar5,&local_2b0,DAT_1803aedd0
                                                                     );
LAB_1800533b8:
                                                        if ((-1 < (int)*plVar7) &&
                                                           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                          lVar5 = FUN_18000cc40(plVar7);
                                                          (**(code **)(lVar5 + 0x30))(plVar7);
                                                        }
                                                        if ((-1 < (int)*plVar9) &&
                                                           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                          lVar5 = FUN_18000cc40(plVar9);
                                                          (**(code **)(lVar5 + 0x30))(plVar9);
                                                        }
                                                        if ((-1 < (int)*plVar6) &&
                                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                          lVar5 = FUN_18000cc40(plVar6);
                                                          (**(code **)(lVar5 + 0x30))(plVar6);
                                                        }
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar5 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar5 + 0x30))(plVar8);
                                                        }
                                                        if ((-1 < (int)*plVar11) &&
                                                           (*plVar11 = *plVar11 + -1, *plVar11 == 0)
                                                           ) {
                                                          lVar5 = FUN_18000cc40(plVar11);
                                                          (**(code **)(lVar5 + 0x30))(plVar11);
                                                        }
                                                        plVar6 = local_260;
                                                        if ((-1 < (int)*plVar12) &&
                                                           (*plVar12 = *plVar12 + -1, *plVar12 == 0)
                                                           ) {
                                                          lVar5 = FUN_18000cc40(plVar12);
                                                          (**(code **)(lVar5 + 0x30))(plVar12);
                                                          plVar6 = local_260;
                                                        }
                                                      }
                                                      else {
                                                        local_250 = (longlong *)
                                                                    FUN_18037f0c0(lVar5,lVar13,
                                                                                  DAT_1803aedd8);
                                                        if (local_250 == (longlong *)0x0) {
                                                          FUN_18000ce30(lVar5,&local_2b0);
                                                          goto LAB_1800533b8;
                                                        }
                                                        local_280 = (longlong *)
                                                                    FUN_180382540(local_2a8,
                                                                                  DAT_1803aee10);
                                                        if (local_280 != (longlong *)0x0) {
                                                          uVar15 = 0x48;
                                                          local_158 = DAT_1803aedf8;
                                                          *(undefined4 *)(local_2a0 + 5) = 0x48;
                                                          local_160 = local_280;
                                                          local_258 = (longlong *)
                                                                      FUN_180377750(local_278,
                                                                                    local_250,
                                                                                    &local_160);
                                                          if ((-1 < (int)*local_250) &&
                                                             (*local_250 = *local_250 + -1,
                                                             *local_250 == 0)) {
                                                            lVar5 = FUN_18000cc40(local_250);
                                                            (**(code **)(lVar5 + 0x30))(local_250);
                                                          }
                                                          if ((-1 < (int)*local_280) &&
                                                             (*local_280 = *local_280 + -1,
                                                             *local_280 == 0)) {
                                                            lVar5 = FUN_18000cc40();
                                                            (**(code **)(lVar5 + 0x30))(local_280);
                                                          }
                                                          lVar5 = local_278;
                                                          if (local_258 == (longlong *)0x0) {
                                                            FUN_18000ce30(local_278,&local_2b0);
                                                            if ((-1 < (int)*plVar7) &&
                                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0
                                                               )) {
                                                              lVar5 = FUN_18000cc40(plVar7);
                                                              (**(code **)(lVar5 + 0x30))(plVar7);
                                                            }
                                                            if ((-1 < (int)*plVar9) &&
                                                               (*plVar9 = *plVar9 + -1, *plVar9 == 0
                                                               )) {
                                                              lVar5 = FUN_18000cc40(plVar9);
                                                              (**(code **)(lVar5 + 0x30))(plVar9);
                                                            }
                                                            if ((-1 < (int)*plVar6) &&
                                                               (*plVar6 = *plVar6 + -1, *plVar6 == 0
                                                               )) {
                                                              lVar5 = FUN_18000cc40(plVar6);
                                                              (**(code **)(lVar5 + 0x30))(plVar6);
                                                            }
                                                            if ((-1 < (int)*plVar8) &&
                                                               (*plVar8 = *plVar8 + -1, *plVar8 == 0
                                                               )) {
                                                              lVar5 = FUN_18000cc40(plVar8);
                                                              (**(code **)(lVar5 + 0x30))(plVar8);
                                                            }
                                                            if ((-1 < (int)*plVar11) &&
                                                               (*plVar11 = *plVar11 + -1,
                                                               *plVar11 == 0)) {
                                                              lVar5 = FUN_18000cc40(plVar11);
                                                              (**(code **)(lVar5 + 0x30))(plVar11);
                                                            }
                                                            plVar7 = local_260;
                                                            if ((-1 < (int)*plVar12) &&
                                                               (*plVar12 = *plVar12 + -1,
                                                               *plVar12 == 0)) {
                                                              lVar5 = FUN_18000cc40(plVar12);
                                                              (**(code **)(lVar5 + 0x30))(plVar12);
                                                              plVar7 = local_260;
                                                            }
LAB_180053708:
                                                            plVar8 = local_2a0;
                                                            plVar16 = local_290;
                                                            param_1 = local_278;
                                                            plVar6 = local_2a8;
                                                            plVar10 = local_298;
                                                            if ((-1 < (int)*plVar7) &&
                                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0
                                                               )) {
                                                              lVar5 = FUN_18000cc40(plVar7);
                                                              (**(code **)(lVar5 + 0x30))(plVar7);
                                                              plVar8 = local_2a0;
                                                              plVar16 = local_290;
                                                              param_1 = local_278;
                                                              plVar6 = local_2a8;
                                                              plVar10 = local_298;
                                                            }
                                                            goto LAB_180052819;
                                                          }
                                                          lVar13 = FUN_180050e30(local_278);
                                                          if (lVar13 == 0) {
                                                            FUN_180373340(lVar5,&local_2b0,
                                                                          DAT_1803aedd0);
LAB_18005375d:
                                                            if ((-1 < (int)*plVar7) &&
                                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0
                                                               )) {
                                                              lVar5 = FUN_18000cc40(plVar7);
                                                              (**(code **)(lVar5 + 0x30))(plVar7);
                                                            }
                                                            if ((-1 < (int)*plVar9) &&
                                                               (*plVar9 = *plVar9 + -1, *plVar9 == 0
                                                               )) {
                                                              lVar5 = FUN_18000cc40(plVar9);
                                                              (**(code **)(lVar5 + 0x30))(plVar9);
                                                            }
                                                            if ((-1 < (int)*plVar6) &&
                                                               (*plVar6 = *plVar6 + -1, *plVar6 == 0
                                                               )) {
                                                              lVar5 = FUN_18000cc40(plVar6);
                                                              (**(code **)(lVar5 + 0x30))(plVar6);
                                                            }
                                                            if ((-1 < (int)*plVar8) &&
                                                               (*plVar8 = *plVar8 + -1, *plVar8 == 0
                                                               )) {
                                                              lVar5 = FUN_18000cc40(plVar8);
                                                              (**(code **)(lVar5 + 0x30))(plVar8);
                                                            }
                                                            if ((-1 < (int)*plVar11) &&
                                                               (*plVar11 = *plVar11 + -1,
                                                               *plVar11 == 0)) {
                                                              lVar5 = FUN_18000cc40(plVar11);
                                                              (**(code **)(lVar5 + 0x30))(plVar11);
                                                            }
                                                            if ((-1 < (int)*plVar12) &&
                                                               (*plVar12 = *plVar12 + -1,
                                                               *plVar12 == 0)) {
                                                              lVar5 = FUN_18000cc40(plVar12);
                                                              (**(code **)(lVar5 + 0x30))(plVar12);
                                                            }
                                                            plVar10 = local_260;
                                                            plVar6 = local_258;
                                                            if ((-1 < (int)*local_260) &&
                                                               (*local_260 = *local_260 + -1,
                                                               *local_260 == 0)) {
                                                              lVar5 = FUN_18000cc40(local_260);
                                                              (**(code **)(lVar5 + 0x30))(plVar10);
                                                              plVar6 = local_258;
                                                            }
                                                          }
                                                          else {
                                                            local_250 = (longlong *)
                                                                        FUN_18037f0c0(lVar5,lVar13,
                                                                                      DAT_1803aedd8)
                                                            ;
                                                            if (local_250 == (longlong *)0x0) {
                                                              FUN_18000ce30(lVar5,&local_2b0);
                                                              goto LAB_18005375d;
                                                            }
                                                            local_2b8 = (longlong *)
                                                                        FUN_180382540(local_2a8,
                                                                                      DAT_1803aee18)
                                                            ;
                                                            if (local_2b8 != (longlong *)0x0) {
                                                              uVar15 = 0x49;
                                                              local_148 = DAT_1803aed90;
                                                              *(undefined4 *)(local_2a0 + 5) = 0x49;
                                                              local_150 = local_2b8;
                                                              local_280 = (longlong *)
                                                                          FUN_180377750(local_278,
                                                                                        local_250,
                                                                                        &local_150);
                                                              if ((-1 < (int)*local_250) &&
                                                                 (*local_250 = *local_250 + -1,
                                                                 *local_250 == 0)) {
                                                                lVar5 = FUN_18000cc40(local_250);
                                                                (**(code **)(lVar5 + 0x30))
                                                                          (local_250);
                                                              }
                                                              if ((-1 < (int)*local_2b8) &&
                                                                 (*local_2b8 = *local_2b8 + -1,
                                                                 *local_2b8 == 0)) {
                                                                lVar5 = FUN_18000cc40();
                                                                (**(code **)(lVar5 + 0x30))
                                                                          (local_2b8);
                                                              }
                                                              if (local_280 == (longlong *)0x0) {
                                                                FUN_18000ce30(local_278,&local_2b0);
                                                                if ((-1 < (int)*plVar7) &&
                                                                   (*plVar7 = *plVar7 + -1,
                                                                   *plVar7 == 0)) {
                                                                  lVar5 = FUN_18000cc40(plVar7);
                                                                  (**(code **)(lVar5 + 0x30))
                                                                            (plVar7);
                                                                }
                                                                if ((-1 < (int)*plVar9) &&
                                                                   (*plVar9 = *plVar9 + -1,
                                                                   *plVar9 == 0)) {
                                                                  lVar5 = FUN_18000cc40(plVar9);
                                                                  (**(code **)(lVar5 + 0x30))
                                                                            (plVar9);
                                                                }
                                                                if ((-1 < (int)*plVar6) &&
                                                                   (*plVar6 = *plVar6 + -1,
                                                                   *plVar6 == 0)) {
                                                                  lVar5 = FUN_18000cc40(plVar6);
                                                                  (**(code **)(lVar5 + 0x30))
                                                                            (plVar6);
                                                                }
                                                                if ((-1 < (int)*plVar8) &&
                                                                   (*plVar8 = *plVar8 + -1,
                                                                   *plVar8 == 0)) {
                                                                  lVar5 = FUN_18000cc40(plVar8);
                                                                  (**(code **)(lVar5 + 0x30))
                                                                            (plVar8);
                                                                }
                                                                if ((-1 < (int)*plVar11) &&
                                                                   (*plVar11 = *plVar11 + -1,
                                                                   *plVar11 == 0)) {
                                                                  lVar5 = FUN_18000cc40(plVar11);
                                                                  (**(code **)(lVar5 + 0x30))
                                                                            (plVar11);
                                                                }
                                                                if ((-1 < (int)*plVar12) &&
                                                                   (*plVar12 = *plVar12 + -1,
                                                                   *plVar12 == 0)) {
                                                                  lVar5 = FUN_18000cc40(plVar12);
                                                                  (**(code **)(lVar5 + 0x30))
                                                                            (plVar12);
                                                                }
                                                                plVar10 = local_260;
                                                                plVar7 = local_258;
                                                                if ((-1 < (int)*local_260) &&
                                                                   (*local_260 = *local_260 + -1,
                                                                   *local_260 == 0)) {
                                                                  lVar5 = FUN_18000cc40(local_260);
                                                                  (**(code **)(lVar5 + 0x30))
                                                                            (plVar10);
                                                                  plVar7 = local_258;
                                                                }
                                                                goto LAB_180053708;
                                                              }
                                                              if (local_298 == (longlong *)0x0) {
                                                                if ((-1 < (int)*plVar7) &&
                                                                   (*plVar7 = *plVar7 + -1,
                                                                   *plVar7 == 0)) {
                                                                  lVar5 = FUN_18000cc40(plVar7);
                                                                  (**(code **)(lVar5 + 0x30))
                                                                            (plVar7);
                                                                }
                                                                if ((-1 < (int)*plVar9) &&
                                                                   (*plVar9 = *plVar9 + -1,
                                                                   *plVar9 == 0)) {
                                                                  lVar5 = FUN_18000cc40(plVar9);
                                                                  (**(code **)(lVar5 + 0x30))
                                                                            (plVar9);
                                                                }
                                                                if ((-1 < (int)*plVar6) &&
                                                                   (*plVar6 = *plVar6 + -1,
                                                                   *plVar6 == 0)) {
                                                                  lVar5 = FUN_18000cc40(plVar6);
                                                                  (**(code **)(lVar5 + 0x30))
                                                                            (plVar6);
                                                                }
                                                                if ((-1 < (int)*plVar8) &&
                                                                   (*plVar8 = *plVar8 + -1,
                                                                   *plVar8 == 0)) {
                                                                  lVar5 = FUN_18000cc40(plVar8);
                                                                  (**(code **)(lVar5 + 0x30))
                                                                            (plVar8);
                                                                }
                                                                if ((-1 < (int)*plVar11) &&
                                                                   (*plVar11 = *plVar11 + -1,
                                                                   *plVar11 == 0)) {
                                                                  lVar5 = FUN_18000cc40(plVar11);
                                                                  (**(code **)(lVar5 + 0x30))
                                                                            (plVar11);
                                                                }
                                                                if ((-1 < (int)*plVar12) &&
                                                                   (*plVar12 = *plVar12 + -1,
                                                                   *plVar12 == 0)) {
                                                                  lVar5 = FUN_18000cc40(plVar12);
                                                                  (**(code **)(lVar5 + 0x30))
                                                                            (plVar12);
                                                                }
                                                                plVar10 = local_260;
                                                                if ((-1 < (int)*local_260) &&
                                                                   (*local_260 = *local_260 + -1,
                                                                   *local_260 == 0)) {
                                                                  lVar5 = FUN_18000cc40(local_260);
                                                                  (**(code **)(lVar5 + 0x30))
                                                                            (plVar10);
                                                                }
                                                                plVar10 = local_258;
                                                                if ((-1 < (int)*local_258) &&
                                                                   (*local_258 = *local_258 + -1,
                                                                   *local_258 == 0)) {
                                                                  lVar5 = FUN_18000cc40(local_258);
                                                                  (**(code **)(lVar5 + 0x30))
                                                                            (plVar10);
                                                                }
                                                                plVar10 = local_280;
                                                                if ((-1 < (int)*local_280) &&
                                                                   (*local_280 = *local_280 + -1,
                                                                   *local_280 == 0)) {
                                                                  lVar5 = FUN_18000cc40(local_280);
                                                                  (**(code **)(lVar5 + 0x30))
                                                                            (plVar10);
                                                                }
                                                                param_1 = local_278;
                                                                FUN_180370fe0(local_278,&local_2b0,
                                                                              DAT_1803aee20);
                                                                FUN_18000d1d0(param_1,local_2b0);
                                                                uVar15 = 0x4a;
                                                                plVar8 = local_2a0;
                                                                plVar16 = local_290;
                                                                plVar6 = local_2a8;
                                                                plVar10 = local_298;
                                                              }
                                                              else {
                                                                local_68 = local_280;
                                                                uVar15 = 0x3e;
                                                                local_60 = local_298;
                                                                *(undefined4 *)(local_2a0 + 5) =
                                                                     0x3e;
                                                                local_b8 = local_238;
                                                                local_b0 = local_208;
                                                                local_78 = local_260;
                                                                local_70 = local_258;
                                                                local_58 = local_1c8;
                                                                local_a8 = plVar7;
                                                                local_a0 = plVar9;
                                                                local_98 = plVar6;
                                                                local_90 = plVar8;
                                                                local_88 = plVar11;
                                                                local_80 = plVar12;
                                                                local_250 = (longlong *)
                                                                            FUN_18037ccb0(local_278,
                                                                                          local_270,
                                                                                          &local_b8,
                                                                                                                                                                                    
                                                  DAT_1803aee28);
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar7);
                                                    (**(code **)(lVar5 + 0x30))(plVar7);
                                                  }
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar5 + 0x30))(plVar9);
                                                  }
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar6);
                                                    (**(code **)(lVar5 + 0x30))(plVar6);
                                                  }
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar8);
                                                    (**(code **)(lVar5 + 0x30))(plVar8);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar5 + 0x30))(plVar11);
                                                  }
                                                  if ((-1 < (int)*plVar12) &&
                                                     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar12);
                                                    (**(code **)(lVar5 + 0x30))(plVar12);
                                                  }
                                                  plVar10 = local_260;
                                                  if ((-1 < (int)*local_260) &&
                                                     (*local_260 = *local_260 + -1, *local_260 == 0)
                                                     ) {
                                                    lVar5 = FUN_18000cc40(local_260);
                                                    (**(code **)(lVar5 + 0x30))(plVar10);
                                                  }
                                                  plVar10 = local_258;
                                                  if ((-1 < (int)*local_258) &&
                                                     (*local_258 = *local_258 + -1, *local_258 == 0)
                                                     ) {
                                                    lVar5 = FUN_18000cc40(local_258);
                                                    (**(code **)(lVar5 + 0x30))(plVar10);
                                                  }
                                                  plVar10 = local_280;
                                                  if ((-1 < (int)*local_280) &&
                                                     (*local_280 = *local_280 + -1, *local_280 == 0)
                                                     ) {
                                                    lVar5 = FUN_18000cc40(local_280);
                                                    (**(code **)(lVar5 + 0x30))(plVar10);
                                                  }
                                                  plVar10 = local_250;
                                                  param_1 = local_278;
                                                  uVar3 = DAT_1803aee38;
                                                  uVar2 = DAT_1803aee30;
                                                  if (local_250 == (longlong *)0x0) {
                                                    FUN_18000ce30(local_278,&local_2b0);
                                                    plVar8 = local_2a0;
                                                    plVar16 = local_290;
                                                    plVar6 = local_2a8;
                                                    plVar10 = local_298;
                                                  }
                                                  else {
                                                    local_1f0 = local_250;
                                                    plVar7 = (longlong *)_PyDict_NewPresized(2);
                                                    PyDict_SetItem(plVar7,uVar2,uVar3);
                                                    param_1 = local_278;
                                                    uVar2 = DAT_1803aee40;
                                                    lVar5 = FUN_180051530(local_278);
                                                    if (lVar5 == 0) {
                                                      FUN_180373340(param_1,&local_2b0,DAT_1803aee48
                                                                   );
                                                      uVar15 = 0x51;
                                                    }
                                                    else {
                                                      uVar15 = 0x51;
                                                      *(undefined4 *)(local_2a0 + 5) = 0x51;
                                                      local_150 = local_198;
                                                      local_148 = local_1a0;
                                                      local_270 = local_290;
                                                      local_358 = (longlong *)DAT_1803aee50;
                                                      plVar6 = (longlong *)
                                                               FUN_18037d650(param_1,lVar5,
                                                                             &local_150,&local_270);
                                                      if (plVar6 != (longlong *)0x0) {
                                                        PyDict_SetItem(plVar7,uVar2,plVar6);
                                                        if ((-1 < (int)*plVar6) &&
                                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                          lVar5 = FUN_18000cc40(plVar6);
                                                          (**(code **)(lVar5 + 0x30))(plVar6);
                                                        }
                                                        plVar9 = (longlong *)
                                                                 FUN_180372160(param_1,2);
                                                        local_260 = plVar9;
                                                        FUN_180024750(plVar9,0,plVar7);
                                                        uVar2 = DAT_1803aee58;
                                                        uVar15 = DAT_1803aee30;
                                                        plVar6 = (longlong *)_PyDict_NewPresized(2);
                                                        PyDict_SetItem(plVar6,uVar15,uVar2);
                                                        uVar15 = DAT_1803aee40;
                                                        lVar5 = FUN_180051640(param_1);
                                                        if (lVar5 == 0) {
                                                          FUN_180373340(param_1,&local_2b0,
                                                                        DAT_1803aee60);
                                                        }
                                                        else {
                                                          *(undefined4 *)(local_2a0 + 5) = 0x53;
                                                          plVar10 = (longlong *)
                                                                    FUN_180376770(param_1,lVar5,
                                                                                  plVar10);
                                                          if (plVar10 != (longlong *)0x0) {
                                                            PyDict_SetItem(plVar6,uVar15,plVar10);
                                                            if ((-1 < (int)*plVar10) &&
                                                               (*plVar10 = *plVar10 + -1,
                                                               *plVar10 == 0)) {
                                                              lVar5 = FUN_18000cc40(plVar10);
                                                              (**(code **)(lVar5 + 0x30))(plVar10);
                                                            }
                                                            FUN_180024750(plVar9,1,plVar6);
                                                            local_1e8 = plVar9;
                                                            plVar10 = (longlong *)
                                                                      FUN_18037f0c0(param_1,
                                                  local_218,DAT_1803aed20);
                                                  if (plVar10 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_2b0);
                                                    uVar15 = 0x55;
                                                    plVar8 = local_2a0;
                                                    plVar16 = local_290;
                                                    plVar6 = local_2a8;
                                                    plVar10 = local_298;
                                                  }
                                                  else {
                                                    plVar6 = (longlong *)
                                                             FUN_18037f0c0(param_1,plVar10,
                                                                           DAT_1803aee68);
                                                    if ((-1 < (int)*plVar10) &&
                                                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                      lVar5 = FUN_18000cc40(plVar10);
                                                      (**(code **)(lVar5 + 0x30))(plVar10);
                                                    }
                                                    if (plVar6 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_2b0);
                                                      uVar15 = 0x55;
                                                      plVar8 = local_2a0;
                                                      plVar16 = local_290;
                                                      plVar6 = local_2a8;
                                                      plVar10 = local_298;
                                                    }
                                                    else {
                                                      plVar10 = (longlong *)FUN_180374a40(param_1,1)
                                                      ;
                                                      iVar4 = (int)*plVar9 + 1;
                                                      if (iVar4 != 0) {
                                                        *(int *)plVar9 = iVar4;
                                                      }
                                                      FUN_18000cc80(plVar10,0,plVar9);
                                                      plVar8 = (longlong *)
                                                               FUN_18036fb80(param_1,DAT_1803aee70);
                                                      uVar15 = 0x55;
                                                      *(undefined4 *)(local_2a0 + 5) = 0x55;
                                                      plVar16 = (longlong *)
                                                                FUN_180001790(param_1,plVar6,plVar10
                                                                              ,plVar8);
                                                      if ((-1 < (int)*plVar6) &&
                                                         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                        lVar5 = FUN_18000cc40(plVar6);
                                                        (**(code **)(lVar5 + 0x30))(plVar6);
                                                      }
                                                      if ((-1 < (int)*plVar10) &&
                                                         (*plVar10 = *plVar10 + -1, *plVar10 == 0))
                                                      {
                                                        lVar5 = FUN_18000cc40(plVar10);
                                                        (**(code **)(lVar5 + 0x30))(plVar10);
                                                      }
                                                      if ((-1 < (int)*plVar8) &&
                                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                        lVar5 = FUN_18000cc40(plVar8);
                                                        (**(code **)(lVar5 + 0x30))(plVar8);
                                                      }
                                                      if (plVar16 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_2b0);
                                                        plVar8 = local_2a0;
                                                        plVar16 = local_290;
                                                        plVar6 = local_2a8;
                                                        plVar10 = local_298;
                                                      }
                                                      else {
                                                        plVar10 = (longlong *)
                                                                  FUN_18037f0c0(param_1,plVar16,
                                                                                DAT_1803aed60);
                                                        if ((-1 < (int)*plVar16) &&
                                                           (*plVar16 = *plVar16 + -1, *plVar16 == 0)
                                                           ) {
                                                          lVar5 = FUN_18000cc40(plVar16);
                                                          (**(code **)(lVar5 + 0x30))(plVar16);
                                                        }
                                                        plVar8 = local_2a0;
                                                        if (plVar10 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_2b0);
                                                          plVar8 = local_2a0;
                                                          plVar16 = local_290;
                                                          plVar6 = local_2a8;
                                                          plVar10 = local_298;
                                                        }
                                                        else {
                                                          *(undefined4 *)(local_2a0 + 5) = 0x55;
                                                          plVar6 = (longlong *)
                                                                   FUN_180376210(param_1,plVar10);
                                                          local_258 = plVar6;
                                                          if ((-1 < (int)*plVar10) &&
                                                             (*plVar10 = *plVar10 + -1,
                                                             *plVar10 == 0)) {
                                                            lVar5 = FUN_18000cc40(plVar10);
                                                            (**(code **)(lVar5 + 0x30))(plVar10);
                                                          }
                                                          if (plVar6 == (longlong *)0x0) {
LAB_180054225:
                                                            FUN_18000ce30(param_1,&local_2b0);
                                                            plVar16 = local_290;
                                                            plVar6 = local_2a8;
                                                            plVar10 = local_298;
                                                          }
                                                          else {
                                                            local_1e0 = plVar6;
                                                            lVar5 = FUN_180051750(param_1);
                                                            if (lVar5 == 0) {
                                                              FUN_180373340(param_1,&local_2b0,
                                                                            DAT_1803aee78);
                                                              uVar15 = 0x57;
                                                              plVar16 = local_290;
                                                              plVar6 = local_2a8;
                                                              plVar10 = local_298;
                                                            }
                                                            else {
                                                              uVar15 = 0x57;
                                                              local_148 = DAT_1803aee80;
                                                              *(undefined4 *)(plVar8 + 5) = 0x57;
                                                              local_150 = plVar6;
                                                              plVar10 = (longlong *)
                                                                        FUN_180377750(param_1,lVar5,
                                                                                      &local_150);
                                                              if (plVar10 == (longlong *)0x0)
                                                              goto LAB_180054225;
                                                              local_248 = plVar10;
                                                              lVar5 = FUN_180051860(param_1);
                                                              if (lVar5 == 0) {
                                                                FUN_180373340(param_1,&local_2b0,
                                                                              DAT_1803aee88);
                                                                uVar15 = 0x58;
                                                                plVar16 = local_290;
                                                                plVar6 = local_2a8;
                                                                plVar10 = local_298;
                                                              }
                                                              else {
                                                                uVar15 = 0x58;
                                                                *(undefined4 *)(plVar8 + 5) = 0x58;
                                                                local_280 = (longlong *)
                                                                            FUN_180376770(param_1,
                                                  lVar5,plVar6);
                                                  if (local_280 == (longlong *)0x0)
                                                  goto LAB_180054225;
                                                  local_1d8 = local_280;
                                                  lVar5 = FUN_180051750(param_1);
                                                  if (lVar5 == 0) {
                                                    FUN_180373340(param_1,&local_2b0,DAT_1803aee78);
                                                    uVar15 = 0x5a;
                                                    plVar16 = local_290;
                                                    plVar6 = local_2a8;
                                                    plVar10 = local_298;
                                                  }
                                                  else {
                                                    uVar15 = 0x5a;
                                                    local_158 = DAT_1803aee90;
                                                    *(undefined4 *)(plVar8 + 5) = 0x5a;
                                                    local_160 = plVar6;
                                                    plVar9 = (longlong *)
                                                             FUN_180377750(param_1,lVar5,&local_160)
                                                    ;
                                                    if (plVar9 == (longlong *)0x0)
                                                    goto LAB_180054225;
                                                    iVar4 = FUN_180001510(plVar9);
                                                    if (iVar4 == -1) {
                                                      FUN_18000ce30(param_1,&local_2b0);
                                                      plVar16 = local_290;
                                                      plVar6 = local_2a8;
                                                      plVar10 = local_298;
                                                      if ((-1 < (int)*plVar9) &&
                                                         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                        lVar5 = FUN_18000cc40(plVar9);
                                                        (**(code **)(lVar5 + 0x30))(plVar9);
                                                        plVar16 = local_290;
                                                        plVar6 = local_2a8;
                                                        plVar10 = local_298;
                                                      }
                                                    }
                                                    else {
                                                      plVar16 = local_1c0;
                                                      if (iVar4 != 1) {
                                                        if ((-1 < (int)*plVar9) &&
                                                           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                          lVar5 = FUN_18000cc40(plVar9);
                                                          (**(code **)(lVar5 + 0x30))(plVar9);
                                                        }
                                                        plVar16 = local_1c0;
                                                        plVar9 = (longlong *)
                                                                 FUN_18037f0c0(param_1,local_1c0,
                                                                               DAT_1803aedc0);
                                                        if (plVar9 == (longlong *)0x0)
                                                        goto LAB_180054225;
                                                      }
                                                      local_230 = plVar9;
                                                      lVar5 = FUN_180051750(param_1);
                                                      if (lVar5 == 0) {
                                                        FUN_180373340(param_1,&local_2b0,
                                                                      DAT_1803aee78);
                                                        uVar15 = 0x5b;
                                                        plVar16 = local_290;
                                                        plVar6 = local_2a8;
                                                        plVar10 = local_298;
                                                      }
                                                      else {
                                                        uVar15 = 0x5b;
                                                        local_168 = DAT_1803aee98;
                                                        *(undefined4 *)(plVar8 + 5) = 0x5b;
                                                        local_170 = plVar6;
                                                        plVar7 = (longlong *)
                                                                 FUN_180377750(param_1,lVar5,
                                                                               &local_170);
                                                        if (plVar7 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_2b0);
                                                          plVar8 = local_2a0;
                                                          plVar16 = local_290;
                                                          plVar6 = local_2a8;
                                                          plVar10 = local_298;
                                                        }
                                                        else {
                                                          iVar4 = FUN_180001510(plVar7);
                                                          if (iVar4 == -1) goto LAB_180053edd;
                                                          if (iVar4 != 1) {
                                                            if ((-1 < (int)*plVar7) &&
                                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0
                                                               )) {
                                                              lVar5 = FUN_18000cc40(plVar7);
                                                              (**(code **)(lVar5 + 0x30))(plVar7);
                                                            }
                                                            plVar7 = (longlong *)
                                                                     FUN_18037f0c0(param_1,plVar16,
                                                                                   DAT_1803aede8);
                                                            if (plVar7 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_2b0);
                                                              plVar8 = local_2a0;
                                                              plVar16 = local_290;
                                                              plVar6 = local_2a8;
                                                              plVar10 = local_298;
                                                              goto LAB_180052819;
                                                            }
                                                          }
                                                          local_228 = plVar7;
                                                          lVar5 = FUN_180051750(param_1);
                                                          if (lVar5 == 0) {
                                                            FUN_180373340(param_1,&local_2b0,
                                                                          DAT_1803aee78);
                                                            uVar15 = 0x5c;
                                                            plVar8 = local_2a0;
                                                            plVar16 = local_290;
                                                            plVar6 = local_2a8;
                                                            plVar10 = local_298;
                                                          }
                                                          else {
                                                            uVar15 = 0x5c;
                                                            local_178 = DAT_1803aeea0;
                                                            *(undefined4 *)(local_2a0 + 5) = 0x5c;
                                                            local_180 = plVar6;
                                                            plVar16 = (longlong *)
                                                                      FUN_180377750(param_1,lVar5,
                                                                                    &local_180);
                                                            plVar8 = local_280;
                                                            if (plVar16 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_2b0);
                                                              plVar8 = local_2a0;
                                                              plVar16 = local_290;
                                                              plVar6 = local_2a8;
                                                              plVar10 = local_298;
                                                            }
                                                            else {
                                                              local_240 = plVar16;
                                                              iVar4 = FUN_180001510(local_280);
                                                              if (iVar4 == -1) {
                                                                FUN_18000ce30(param_1,&local_2b0);
                                                                uVar15 = 0x5e;
                                                                plVar8 = local_2a0;
                                                                plVar16 = local_290;
                                                                plVar6 = local_2a8;
                                                                plVar10 = local_298;
                                                              }
                                                              else if (iVar4 == 0) {
                                                                lVar5 = FUN_180051200(param_1);
                                                                plVar7 = local_1c0;
                                                                if (lVar5 == 0) {
                                                                  FUN_180373340(param_1,&local_2b0,
                                                                                DAT_1803aeee8);
                                                                  uVar15 = 0x6a;
                                                                  plVar8 = local_2a0;
                                                                  plVar16 = local_290;
                                                                  plVar6 = local_2a8;
                                                                  plVar10 = local_298;
                                                                }
                                                                else {
                                                                  plVar10 = (longlong *)
                                                                            FUN_18037f0c0(param_1,
                                                  local_1c0,DAT_1803aedf0);
                                                  if (plVar10 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_2b0);
                                                    uVar15 = 0x6b;
                                                    plVar8 = local_2a0;
                                                    plVar16 = local_290;
                                                    plVar6 = local_2a8;
                                                    plVar10 = local_298;
                                                  }
                                                  else {
                                                    uVar15 = 0x6a;
                                                    *(undefined4 *)(local_2a0 + 5) = 0x6a;
                                                    local_130 = local_238;
                                                    local_128 = local_208;
                                                    local_138 = plVar10;
                                                    local_120 = plVar16;
                                                    plVar8 = (longlong *)
                                                             FUN_180379a80(param_1,lVar5,&local_138)
                                                    ;
                                                    if ((-1 < (int)*plVar10) &&
                                                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                      lVar5 = FUN_18000cc40(plVar10);
                                                      (**(code **)(lVar5 + 0x30))(plVar10);
                                                    }
                                                    if (plVar8 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_2b0);
                                                      plVar8 = local_2a0;
                                                      plVar16 = local_290;
                                                      plVar6 = local_2a8;
                                                      plVar10 = local_298;
                                                    }
                                                    else {
LAB_180054a02:
                                                      local_220 = plVar8;
                                                      lVar5 = FUN_180051ca0(param_1);
                                                      plVar8 = local_2a0;
                                                      if (lVar5 == 0) {
                                                        FUN_180373340(param_1,&local_2b0,
                                                                      DAT_1803aeef0);
                                                        uVar15 = 0x70;
                                                        plVar8 = local_2a0;
                                                        plVar16 = local_290;
                                                        plVar6 = local_2a8;
                                                        plVar10 = local_298;
                                                      }
                                                      else {
                                                        uVar15 = 0x70;
                                                        local_1c0 = local_2a8;
                                                        *(undefined4 *)(local_2a0 + 5) = 0x70;
                                                        local_1b8 = plVar6;
                                                        plVar10 = (longlong *)
                                                                  FUN_180377750(param_1,lVar5,
                                                                                &local_1c0);
                                                        if (plVar10 == (longlong *)0x0)
                                                        goto LAB_180054225;
                                                        if ((-1 < (int)*plVar10) &&
                                                           (*plVar10 = *plVar10 + -1, *plVar10 == 0)
                                                           ) {
                                                          lVar5 = FUN_18000cc40(plVar10);
                                                          (**(code **)(lVar5 + 0x30))(plVar10);
                                                        }
                                                        lVar5 = FUN_180051050(param_1);
                                                        plVar10 = local_298;
                                                        if (lVar5 == 0) {
                                                          FUN_180373340(param_1,&local_2b0,
                                                                        DAT_1803aeef8);
                                                          uVar15 = 0x71;
                                                          plVar16 = local_290;
                                                          plVar6 = local_2a8;
                                                          plVar10 = local_298;
                                                        }
                                                        else if (plVar16 == (longlong *)0x0) {
                                                          FUN_180370fe0(param_1,&local_2b0,
                                                                        DAT_1803aef00);
                                                          FUN_18000d1d0(param_1,local_2b0);
                                                          uVar15 = 0x74;
                                                          plVar16 = local_290;
                                                          plVar6 = local_2a8;
                                                          plVar10 = local_298;
                                                        }
                                                        else {
                                                          local_e8 = local_230;
                                                          uVar15 = 0x71;
                                                          local_e0 = local_228;
                                                          local_d0 = local_220;
                                                          local_c8 = local_248;
                                                          *(undefined4 *)(plVar8 + 5) = 0x71;
                                                          local_c0 = local_298;
                                                          local_d8 = plVar16;
                                                          lVar5 = FUN_18037ccb0(param_1,lVar5,
                                                                                &local_e8,
                                                                                DAT_1803aef08);
                                                          if (lVar5 != 0) {
                                                            FUN_180001d80(param_1);
                                                            if ((-1 < (int)*plVar10) &&
                                                               (*plVar10 = *plVar10 + -1,
                                                               *plVar10 == 0)) {
                                                              lVar13 = FUN_18000cc40(plVar10);
                                                              (**(code **)(lVar13 + 0x30))(plVar10);
                                                            }
                                                            if ((-1 < (int)*plVar7) &&
                                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0
                                                               )) {
                                                              lVar13 = FUN_18000cc40(local_1b0);
                                                              (**(code **)(lVar13 + 0x30))
                                                                        (local_1b0);
                                                            }
                                                            if ((-1 < (int)*local_1a0) &&
                                                               (*local_1a0 = *local_1a0 + -1,
                                                               *local_1a0 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_1a8);
                                                              (**(code **)(lVar13 + 0x30))
                                                                        (local_1a8);
                                                            }
                                                            if ((-1 < (int)*local_198) &&
                                                               (*local_198 = *local_198 + -1,
                                                               *local_198 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_1f8);
                                                              (**(code **)(lVar13 + 0x30))
                                                                        (local_1f8);
                                                            }
                                                            plVar10 = local_290;
                                                            if ((-1 < (int)*local_290) &&
                                                               (*local_290 = *local_290 + -1,
                                                               *local_290 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_290);
                                                              (**(code **)(lVar13 + 0x30))(plVar10);
                                                            }
                                                            if ((-1 < (int)*local_250) &&
                                                               (*local_250 = *local_250 + -1,
                                                               *local_250 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_1f0);
                                                              (**(code **)(lVar13 + 0x30))
                                                                        (local_1f0);
                                                            }
                                                            if ((-1 < (int)*local_260) &&
                                                               (*local_260 = *local_260 + -1,
                                                               *local_260 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_1e8);
                                                              (**(code **)(lVar13 + 0x30))
                                                                        (local_1e8);
                                                            }
                                                            if ((-1 < (int)*plVar6) &&
                                                               (*plVar6 = *plVar6 + -1, *plVar6 == 0
                                                               )) {
                                                              lVar13 = FUN_18000cc40(local_1e0);
                                                              (**(code **)(lVar13 + 0x30))
                                                                        (local_1e0);
                                                            }
                                                            plVar10 = local_248;
                                                            if ((-1 < (int)*local_248) &&
                                                               (*local_248 = *local_248 + -1,
                                                               *local_248 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_248);
                                                              (**(code **)(lVar13 + 0x30))(plVar10);
                                                            }
                                                            if ((-1 < (int)*local_280) &&
                                                               (*local_280 = *local_280 + -1,
                                                               *local_280 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_1d8);
                                                              (**(code **)(lVar13 + 0x30))
                                                                        (local_1d8);
                                                            }
                                                            plVar10 = local_230;
                                                            if ((-1 < (int)*local_230) &&
                                                               (*local_230 = *local_230 + -1,
                                                               *local_230 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_230);
                                                              (**(code **)(lVar13 + 0x30))(plVar10);
                                                            }
                                                            plVar10 = local_228;
                                                            if ((-1 < (int)*local_228) &&
                                                               (*local_228 = *local_228 + -1,
                                                               *local_228 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_228);
                                                              (**(code **)(lVar13 + 0x30))(plVar10);
                                                            }
                                                            plVar10 = local_240;
                                                            if ((-1 < (int)*local_240) &&
                                                               (*local_240 = *local_240 + -1,
                                                               *local_240 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_240);
                                                              (**(code **)(lVar13 + 0x30))(plVar10);
                                                            }
                                                            plVar10 = local_1d0;
                                                            if (((local_1d0 != (longlong *)0x0) &&
                                                                (-1 < (int)*local_1d0)) &&
                                                               (*local_1d0 = *local_1d0 + -1,
                                                               *local_1d0 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_1d0);
                                                              (**(code **)(lVar13 + 0x30))(plVar10);
                                                            }
                                                            plVar10 = local_220;
                                                            if ((-1 < (int)*local_220) &&
                                                               (*local_220 = *local_220 + -1,
                                                               *local_220 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_220);
                                                              (**(code **)(lVar13 + 0x30))(plVar10);
                                                            }
                                                            plVar10 = local_218;
                                                            if ((-1 < (int)*local_218) &&
                                                               (*local_218 = *local_218 + -1,
                                                               *local_218 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_218);
                                                              (**(code **)(lVar13 + 0x30))(plVar10);
                                                            }
                                                            plVar10 = local_210;
                                                            if ((-1 < (int)*local_210) &&
                                                               (*local_210 = *local_210 + -1,
                                                               *local_210 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_210);
                                                              (**(code **)(lVar13 + 0x30))(plVar10);
                                                            }
                                                            plVar10 = local_2a8;
                                                            if ((-1 < (int)*local_2a8) &&
                                                               (*local_2a8 = *local_2a8 + -1,
                                                               *local_2a8 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_2a8);
                                                              (**(code **)(lVar13 + 0x30))(plVar10);
                                                            }
                                                            plVar10 = local_238;
                                                            if ((-1 < (int)*local_238) &&
                                                               (*local_238 = *local_238 + -1,
                                                               *local_238 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_238);
                                                              (**(code **)(lVar13 + 0x30))(plVar10);
                                                            }
                                                            plVar10 = local_208;
                                                            if ((-1 < (int)*local_208) &&
                                                               (*local_208 = *local_208 + -1,
                                                               *local_208 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_208);
                                                              (**(code **)(lVar13 + 0x30))(plVar10);
                                                            }
                                                            plVar10 = local_200;
                                                            if ((-1 < (int)*local_200) &&
                                                               (*local_200 = *local_200 + -1,
                                                               *local_200 == 0)) {
                                                              lVar13 = FUN_18000cc40(local_200);
                                                              (**(code **)(lVar13 + 0x30))(plVar10);
                                                            }
                                                            plVar10 = local_1c8;
                                                            if ((int)*local_1c8 < 0) {
                                                              return lVar5;
                                                            }
                                                            *local_1c8 = *local_1c8 + -1;
                                                            if (*local_1c8 != 0) {
                                                              return lVar5;
                                                            }
                                                            lVar13 = FUN_18000cc40(local_1c8);
                                                            (**(code **)(lVar13 + 0x30))(plVar10);
                                                            return lVar5;
                                                          }
                                                          FUN_18000ce30(param_1,&local_2b0);
                                                          plVar16 = local_290;
                                                          plVar6 = local_2a8;
                                                        }
                                                      }
                                                    }
                                                  }
                                                  }
                                                  }
                                                  else {
                                                    iVar4 = FUN_180001510(plVar10);
                                                    if (iVar4 == -1) {
LAB_1800545af:
                                                      FUN_18000ce30(param_1,&local_2b0);
                                                      uVar15 = 0x5f;
                                                      plVar8 = local_2a0;
                                                      plVar16 = local_290;
                                                      plVar6 = local_2a8;
                                                      plVar10 = local_298;
                                                    }
                                                    else {
                                                      if (iVar4 == 1) {
                                                        iVar4 = (int)*plVar10 + 1;
                                                        plVar6 = plVar10;
                                                        if (iVar4 != 0) {
                                                          *(int *)plVar10 = iVar4;
                                                        }
                                                      }
                                                      else {
                                                        plVar6 = (longlong *)
                                                                 FUN_18037f0c0(param_1,plVar8,
                                                                               DAT_1803aeea8);
                                                        if (plVar6 == (longlong *)0x0)
                                                        goto LAB_1800545af;
                                                      }
                                                      local_248 = plVar6;
                                                      if ((-1 < (int)*plVar10) &&
                                                         (*plVar10 = *plVar10 + -1, *plVar10 == 0))
                                                      {
                                                        lVar5 = FUN_18000cc40(plVar10);
                                                        (**(code **)(lVar5 + 0x30))(plVar10);
                                                      }
                                                      iVar4 = FUN_180001510(plVar16);
                                                      if (iVar4 == -1) {
                                                        FUN_18000ce30(param_1,&local_2b0);
                                                        uVar15 = 0x60;
                                                        plVar8 = local_2a0;
                                                        plVar16 = local_290;
                                                        plVar6 = local_2a8;
                                                        plVar10 = local_298;
                                                      }
                                                      else if (iVar4 == 1) {
                                                        iVar4 = (int)*plVar16 + 1;
                                                        plVar6 = plVar16;
                                                        if (iVar4 != 0) {
                                                          *(int *)plVar16 = iVar4;
                                                        }
LAB_1800546fa:
                                                        local_240 = plVar6;
                                                        if ((-1 < (int)*plVar16) &&
                                                           (*plVar16 = *plVar16 + -1, *plVar16 == 0)
                                                           ) {
                                                          lVar5 = FUN_18000cc40(plVar16);
                                                          (**(code **)(lVar5 + 0x30))(plVar16);
                                                        }
                                                        lVar5 = FUN_180051310(param_1);
                                                        if (lVar5 == 0) {
                                                          FUN_180373340(param_1,&local_2b0,
                                                                        DAT_1803aeec0);
                                                          uVar15 = 0x61;
                                                          plVar8 = local_2a0;
                                                          plVar16 = local_290;
                                                          plVar6 = local_2a8;
                                                          plVar10 = local_298;
                                                        }
                                                        else {
                                                          plVar10 = (longlong *)
                                                                    FUN_18037f0c0(param_1,local_2a8,
                                                                                  DAT_1803aeec8);
                                                          if (plVar10 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_2b0);
                                                            uVar15 = 0x62;
                                                            plVar8 = local_2a0;
                                                            plVar16 = local_290;
                                                            plVar6 = local_2a8;
                                                            plVar10 = local_298;
                                                          }
                                                          else {
                                                            plVar6 = (longlong *)
                                                                     FUN_18037f0c0(param_1,plVar8,
                                                                                   DAT_1803aeeb8);
                                                            if (plVar6 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_2b0);
                                                              if ((-1 < (int)*plVar10) &&
                                                                 (*plVar10 = *plVar10 + -1,
                                                                 *plVar10 == 0)) {
                                                                lVar5 = FUN_18000cc40(plVar10);
                                                                (**(code **)(lVar5 + 0x30))(plVar10)
                                                                ;
                                                              }
                                                              uVar15 = 99;
                                                              plVar8 = local_2a0;
                                                              plVar16 = local_290;
                                                              plVar6 = local_2a8;
                                                              plVar10 = local_298;
                                                            }
                                                            else {
                                                              uVar15 = 0x61;
                                                              *(undefined4 *)(local_2a0 + 5) = 0x61;
                                                              local_270 = local_290;
                                                              local_358 = (longlong *)DAT_1803aee50;
                                                              local_190 = plVar10;
                                                              local_188 = plVar6;
                                                              plVar8 = (longlong *)
                                                                       FUN_18037d650(param_1,lVar5,
                                                                                     &local_190,
                                                                                     &local_270);
                                                              if ((-1 < (int)*plVar10) &&
                                                                 (*plVar10 = *plVar10 + -1,
                                                                 *plVar10 == 0)) {
                                                                lVar5 = FUN_18000cc40(plVar10);
                                                                (**(code **)(lVar5 + 0x30))(plVar10)
                                                                ;
                                                              }
                                                              if ((-1 < (int)*plVar6) &&
                                                                 (*plVar6 = *plVar6 + -1,
                                                                 *plVar6 == 0)) {
                                                                lVar5 = FUN_18000cc40(plVar6);
                                                                (**(code **)(lVar5 + 0x30))(plVar6);
                                                              }
                                                              if (plVar8 == (longlong *)0x0) {
                                                                FUN_18000ce30(param_1,&local_2b0);
                                                                plVar8 = local_2a0;
                                                                plVar16 = local_290;
                                                                plVar6 = local_2a8;
                                                                plVar10 = local_298;
                                                              }
                                                              else {
                                                                local_1d0 = plVar8;
                                                                plVar10 = (longlong *)
                                                                          FUN_18037f0c0(param_1,
                                                  plVar8,DAT_1803aeed0);
                                                  if (plVar10 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_2b0);
                                                    uVar15 = 0x66;
                                                    plVar8 = local_2a0;
                                                    plVar16 = local_290;
                                                    plVar6 = local_2a8;
                                                    plVar10 = local_298;
                                                  }
                                                  else {
                                                    local_230 = plVar10;
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      lVar5 = FUN_18000cc40(plVar9);
                                                      (**(code **)(lVar5 + 0x30))(plVar9);
                                                    }
                                                    plVar10 = (longlong *)
                                                              FUN_18037f0c0(param_1,plVar8,
                                                                            DAT_1803aeed8);
                                                    if (plVar10 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_2b0);
                                                      uVar15 = 0x67;
                                                      plVar8 = local_2a0;
                                                      plVar16 = local_290;
                                                      plVar6 = local_2a8;
                                                      plVar10 = local_298;
                                                    }
                                                    else {
                                                      local_228 = plVar10;
                                                      if ((-1 < (int)*plVar7) &&
                                                         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                        lVar5 = FUN_18000cc40(plVar7);
                                                        (**(code **)(lVar5 + 0x30))(plVar7);
                                                      }
                                                      plVar8 = (longlong *)
                                                               FUN_18037f0c0(param_1,plVar8,
                                                                             DAT_1803aeee0);
                                                      plVar6 = local_258;
                                                      plVar7 = local_1c0;
                                                      plVar16 = local_240;
                                                      if (plVar8 != (longlong *)0x0)
                                                      goto LAB_180054a02;
                                                      FUN_18000ce30(param_1,&local_2b0);
                                                      uVar15 = 0x68;
                                                      plVar8 = local_2a0;
                                                      plVar16 = local_290;
                                                      plVar6 = local_2a8;
                                                      plVar10 = local_298;
                                                    }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  else {
                                                    lVar5 = FUN_180051b90(param_1);
                                                    if (lVar5 == 0) {
                                                      FUN_180373340(param_1,&local_2b0,DAT_1803aeeb0
                                                                   );
                                                      uVar15 = 0x60;
                                                      plVar8 = local_2a0;
                                                      plVar16 = local_290;
                                                      plVar6 = local_2a8;
                                                      plVar10 = local_298;
                                                    }
                                                    else {
                                                      plVar10 = (longlong *)
                                                                FUN_18037f0c0(param_1,plVar8,
                                                                              DAT_1803aeeb8);
                                                      if (plVar10 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_2b0);
                                                        uVar15 = 0x60;
                                                        plVar8 = local_2a0;
                                                        plVar16 = local_290;
                                                        plVar6 = local_2a8;
                                                        plVar10 = local_298;
                                                      }
                                                      else {
                                                        uVar15 = 0x60;
                                                        *(undefined4 *)(local_2a0 + 5) = 0x60;
                                                        plVar6 = (longlong *)
                                                                 FUN_180376770(param_1,lVar5,plVar10
                                                                              );
                                                        if ((-1 < (int)*plVar10) &&
                                                           (*plVar10 = *plVar10 + -1, *plVar10 == 0)
                                                           ) {
                                                          lVar5 = FUN_18000cc40(plVar10);
                                                          (**(code **)(lVar5 + 0x30))(plVar10);
                                                        }
                                                        plVar8 = local_280;
                                                        if (plVar6 != (longlong *)0x0)
                                                        goto LAB_1800546fa;
                                                        FUN_18000ce30(param_1,&local_2b0);
                                                        plVar8 = local_2a0;
                                                        plVar16 = local_290;
                                                        plVar6 = local_2a8;
                                                        plVar10 = local_298;
                                                      }
                                                    }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  goto LAB_180052819;
                                                  }
                                                  FUN_18000ce30(param_1,&local_2b0);
                                                  }
                                                  uVar15 = 0x53;
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar6);
                                                    (**(code **)(lVar5 + 0x30))(plVar6);
                                                  }
                                                  plVar8 = local_2a0;
                                                  plVar16 = local_290;
                                                  plVar6 = local_2a8;
                                                  plVar10 = local_298;
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar5 + 0x30))(plVar9);
                                                    plVar8 = local_2a0;
                                                    plVar16 = local_290;
                                                    plVar6 = local_2a8;
                                                    plVar10 = local_298;
                                                  }
                                                  goto LAB_180052819;
                                                  }
LAB_180053edd:
                                                  FUN_18000ce30(param_1,&local_2b0);
                                                  }
                                                  plVar8 = local_2a0;
                                                  plVar16 = local_290;
                                                  plVar6 = local_2a8;
                                                  plVar10 = local_298;
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar7);
                                                    (**(code **)(lVar5 + 0x30))(plVar7);
                                                    plVar8 = local_2a0;
                                                    plVar16 = local_290;
                                                    plVar6 = local_2a8;
                                                    plVar10 = local_298;
                                                  }
                                                  }
                                                  }
                                                  goto LAB_180052819;
                                                  }
                                                  FUN_18000ce30(lVar5,&local_2b0);
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar7);
                                                    (**(code **)(lVar5 + 0x30))(plVar7);
                                                  }
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar5 + 0x30))(plVar9);
                                                  }
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar6);
                                                    (**(code **)(lVar5 + 0x30))(plVar6);
                                                  }
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar8);
                                                    (**(code **)(lVar5 + 0x30))(plVar8);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar5 + 0x30))(plVar11);
                                                  }
                                                  if ((-1 < (int)*plVar12) &&
                                                     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar12);
                                                    (**(code **)(lVar5 + 0x30))(plVar12);
                                                  }
                                                  plVar10 = local_260;
                                                  if ((-1 < (int)*local_260) &&
                                                     (*local_260 = *local_260 + -1, *local_260 == 0)
                                                     ) {
                                                    lVar5 = FUN_18000cc40(local_260);
                                                    (**(code **)(lVar5 + 0x30))(plVar10);
                                                  }
                                                  plVar10 = local_258;
                                                  plVar6 = local_250;
                                                  if ((-1 < (int)*local_258) &&
                                                     (*local_258 = *local_258 + -1, *local_258 == 0)
                                                     ) {
                                                    lVar5 = FUN_18000cc40(local_258);
                                                    (**(code **)(lVar5 + 0x30))(plVar10);
                                                    plVar6 = local_250;
                                                  }
                                                  }
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar6);
                                                    (**(code **)(lVar5 + 0x30))(plVar6);
                                                  }
                                                  uVar15 = 0x49;
                                                  plVar8 = local_2a0;
                                                  plVar16 = local_290;
                                                  param_1 = local_278;
                                                  plVar6 = local_2a8;
                                                  plVar10 = local_298;
                                                  goto LAB_180052819;
                                                  }
                                                  FUN_18000ce30(lVar5,&local_2b0);
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar7);
                                                    (**(code **)(lVar5 + 0x30))(plVar7);
                                                  }
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar5 + 0x30))(plVar9);
                                                  }
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar6);
                                                    (**(code **)(lVar5 + 0x30))(plVar6);
                                                  }
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar8);
                                                    (**(code **)(lVar5 + 0x30))(plVar8);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar5 + 0x30))(plVar11);
                                                  }
                                                  if ((-1 < (int)*plVar12) &&
                                                     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar12);
                                                    (**(code **)(lVar5 + 0x30))(plVar12);
                                                  }
                                                  plVar10 = local_260;
                                                  plVar6 = local_250;
                                                  if ((-1 < (int)*local_260) &&
                                                     (*local_260 = *local_260 + -1, *local_260 == 0)
                                                     ) {
                                                    lVar5 = FUN_18000cc40(local_260);
                                                    (**(code **)(lVar5 + 0x30))(plVar10);
                                                    plVar6 = local_250;
                                                  }
                                                  }
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar6);
                                                    (**(code **)(lVar5 + 0x30))(plVar6);
                                                  }
                                                  uVar15 = 0x48;
                                                  plVar8 = local_2a0;
                                                  plVar16 = local_290;
                                                  param_1 = local_278;
                                                  plVar6 = local_2a8;
                                                  plVar10 = local_298;
                                                  goto LAB_180052819;
                                                  }
                                                  FUN_18000ce30(param_1,&local_2b0);
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar7);
                                                    (**(code **)(lVar5 + 0x30))(plVar7);
                                                  }
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar5 + 0x30))(plVar9);
                                                  }
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar6);
                                                    (**(code **)(lVar5 + 0x30))(plVar6);
                                                  }
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar8);
                                                    (**(code **)(lVar5 + 0x30))(plVar8);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar5 + 0x30))(plVar11);
                                                  }
                                                  if ((-1 < (int)*plVar12) &&
                                                     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar12);
                                                    (**(code **)(lVar5 + 0x30))(plVar12);
                                                  }
                                                  plVar10 = local_258;
                                                  if ((-1 < (int)*local_258) &&
                                                     (*local_258 = *local_258 + -1, *local_258 == 0)
                                                     ) {
                                                    lVar5 = FUN_18000cc40(local_258);
                                                    (**(code **)(lVar5 + 0x30))(plVar10);
                                                    uVar15 = 0x47;
                                                    plVar8 = local_2a0;
                                                    plVar16 = local_290;
                                                    param_1 = local_278;
                                                    plVar6 = local_2a8;
                                                    plVar10 = local_298;
                                                    goto LAB_180052819;
                                                  }
                                                  }
                                                  uVar15 = 0x47;
                                                  plVar8 = local_2a0;
                                                  plVar16 = local_290;
                                                  param_1 = local_278;
                                                  plVar6 = local_2a8;
                                                  plVar10 = local_298;
                                                  goto LAB_180052819;
                                                  }
                                                  FUN_18000ce30(param_1,&local_2b0);
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar7);
                                                    (**(code **)(lVar5 + 0x30))(plVar7);
                                                  }
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar5 + 0x30))(plVar9);
                                                  }
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar6);
                                                    (**(code **)(lVar5 + 0x30))(plVar6);
                                                  }
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar8);
                                                    (**(code **)(lVar5 + 0x30))(plVar8);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar5 + 0x30))(plVar11);
                                                  }
                                                  if ((-1 < (int)*plVar10) &&
                                                     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar10);
                                                    (**(code **)(lVar5 + 0x30))(plVar10);
                                                    uVar15 = 0x46;
                                                    plVar8 = local_2a0;
                                                    plVar16 = local_290;
                                                    plVar6 = local_2a8;
                                                    plVar10 = local_298;
                                                    goto LAB_180052819;
                                                  }
                                                  }
                                                  uVar15 = 0x46;
                                                  plVar8 = local_2a0;
                                                  plVar16 = local_290;
                                                  plVar6 = local_2a8;
                                                  plVar10 = local_298;
                                                  goto LAB_180052819;
                                                  }
                                                  FUN_18000ce30(param_1,&local_2b0);
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar7);
                                                    (**(code **)(lVar5 + 0x30))(plVar7);
                                                  }
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar5 + 0x30))(plVar9);
                                                  }
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar6);
                                                    (**(code **)(lVar5 + 0x30))(plVar6);
                                                  }
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar8);
                                                    (**(code **)(lVar5 + 0x30))(plVar8);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar5 + 0x30))(plVar11);
                                                    uVar15 = 0x45;
                                                    plVar8 = local_2a0;
                                                    plVar16 = local_290;
                                                    plVar6 = local_2a8;
                                                    plVar10 = local_298;
                                                    goto LAB_180052819;
                                                  }
                                                  }
                                                  uVar15 = 0x45;
                                                  plVar8 = local_2a0;
                                                  plVar16 = local_290;
                                                  plVar6 = local_2a8;
                                                  plVar10 = local_298;
                                                  goto LAB_180052819;
                                                  }
                                                  FUN_18000ce30(param_1,&local_2b0);
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar7);
                                                    (**(code **)(lVar5 + 0x30))(plVar7);
                                                  }
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar5 + 0x30))(plVar9);
                                                  }
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar6);
                                                    (**(code **)(lVar5 + 0x30))(plVar6);
                                                  }
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    lVar5 = FUN_18000cc40(plVar8);
                                                    (**(code **)(lVar5 + 0x30))(plVar8);
                                                    uVar15 = 0x44;
                                                    plVar8 = local_2a0;
                                                    plVar16 = local_290;
                                                    plVar6 = local_2a8;
                                                    plVar10 = local_298;
                                                    goto LAB_180052819;
                                                  }
                                                  }
                                                  uVar15 = 0x44;
                                                  plVar8 = local_2a0;
                                                  plVar16 = local_290;
                                                  plVar6 = local_2a8;
                                                  plVar10 = local_298;
                                                  goto LAB_180052819;
                                                }
                                                local_2b0 = *(longlong *)(param_1 + 0x60);
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                              }
                                              if ((-1 < (int)*plVar7) &&
                                                 (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                (**(code **)(plVar7[1] + 0x30))(plVar7);
                                              }
                                              if ((-1 < (int)*plVar9) &&
                                                 (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                (**(code **)(plVar9[1] + 0x30))(plVar9);
                                              }
                                              uVar15 = 0x43;
                                              plVar8 = local_2a0;
                                              plVar16 = local_290;
                                              plVar6 = local_2a8;
                                              plVar10 = local_298;
                                              goto LAB_180052819;
                                            }
                                            local_2b0 = *(longlong *)(param_1 + 0x60);
                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                            if ((-1 < (int)*plVar7) &&
                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                              (**(code **)(plVar7[1] + 0x30))(plVar7);
                                            }
                                            if ((-1 < (int)*plVar9) &&
                                               (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                              pcVar14 = *(code **)(plVar9[1] + 0x30);
                                              goto LAB_180052802;
                                            }
                                          }
                                          uVar15 = 0x42;
                                          plVar8 = local_2a0;
                                          plVar16 = local_290;
                                          plVar6 = local_2a8;
                                          plVar10 = local_298;
                                        }
                                      }
                                    }
                                  }
                                }
                                goto LAB_180052819;
                              }
                            }
                            uVar15 = 0x37;
                            plVar10 = plVar9;
                            goto LAB_180052413;
                          }
                        }
                        local_2b0 = *(longlong *)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        plVar8 = local_2a0;
                        plVar6 = local_2a8;
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
LAB_180052819:
  piVar1 = *(int **)(local_2b0 + 0x28);
  if (piVar1 == (int *)0x0) {
    lVar5 = FUN_18038e2a0(plVar8,uVar15);
  }
  else {
    if (*(longlong **)(piVar1 + 6) == plVar8) goto LAB_180054b9f;
    lVar5 = FUN_18038e2a0(plVar8,uVar15);
    *(int **)(lVar5 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
  }
  FUN_180001300(&local_2b0,lVar5);
LAB_180054b9f:
  plVar9 = local_240;
  local_2c0 = local_220;
  local_2c8 = local_1d0;
  local_2d0 = local_240;
  local_2d8 = local_228;
  local_2e0 = local_230;
  local_2e8 = local_1d8;
  local_2f0 = local_248;
  local_2f8 = local_1e0;
  local_300 = local_1e8;
  local_308 = local_1f0;
  local_318 = local_1f8;
  local_320 = local_1a8;
  local_328 = local_1b0;
  local_330 = local_1c8;
  local_340 = local_200;
  local_348 = local_208;
  local_350 = local_238;
  local_358 = plVar6;
  local_338 = plVar10;
  local_310 = plVar16;
  FUN_18036d0a0(plVar8,"oooooooooooooooooooooo",local_218,local_210);
  plVar7 = DAT_1803aec80;
  if (plVar8 == DAT_1803aec80) {
    if ((-1 < (int)*DAT_1803aec80) && (*DAT_1803aec80 = *DAT_1803aec80 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    DAT_1803aec80 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  lVar5 = local_2b0;
  local_2b0 = 0;
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))(plVar10);
  }
  if (((local_1b0 != (longlong *)0x0) && (-1 < (int)*local_1b0)) &&
     (*local_1b0 = *local_1b0 + -1, *local_1b0 == 0)) {
    (**(code **)(local_1b0[1] + 0x30))();
  }
  if (((local_1a8 != (longlong *)0x0) && (-1 < (int)*local_1a8)) &&
     (*local_1a8 = *local_1a8 + -1, *local_1a8 == 0)) {
    (**(code **)(local_1a8[1] + 0x30))();
  }
  if (((local_1f8 != (longlong *)0x0) && (-1 < (int)*local_1f8)) &&
     (*local_1f8 = *local_1f8 + -1, *local_1f8 == 0)) {
    (**(code **)(local_1f8[1] + 0x30))();
  }
  if (((plVar16 != (longlong *)0x0) && (-1 < (int)*plVar16)) &&
     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
    (**(code **)(plVar16[1] + 0x30))(plVar16);
  }
  if (((local_1f0 != (longlong *)0x0) && (-1 < (int)*local_1f0)) &&
     (*local_1f0 = *local_1f0 + -1, *local_1f0 == 0)) {
    (**(code **)(local_1f0[1] + 0x30))();
  }
  if (((local_1e8 != (longlong *)0x0) && (-1 < (int)*local_1e8)) &&
     (*local_1e8 = *local_1e8 + -1, *local_1e8 == 0)) {
    (**(code **)(local_1e8[1] + 0x30))();
  }
  if (((local_1e0 != (longlong *)0x0) && (-1 < (int)*local_1e0)) &&
     (*local_1e0 = *local_1e0 + -1, *local_1e0 == 0)) {
    (**(code **)(local_1e0[1] + 0x30))();
  }
  if (((local_248 != (longlong *)0x0) && (-1 < (int)*local_248)) &&
     (*local_248 = *local_248 + -1, *local_248 == 0)) {
    (**(code **)(local_248[1] + 0x30))();
  }
  if (((local_1d8 != (longlong *)0x0) && (-1 < (int)*local_1d8)) &&
     (*local_1d8 = *local_1d8 + -1, *local_1d8 == 0)) {
    (**(code **)(local_1d8[1] + 0x30))();
  }
  if (((local_230 != (longlong *)0x0) && (-1 < (int)*local_230)) &&
     (*local_230 = *local_230 + -1, *local_230 == 0)) {
    (**(code **)(local_230[1] + 0x30))();
  }
  if (((local_228 != (longlong *)0x0) && (-1 < (int)*local_228)) &&
     (*local_228 = *local_228 + -1, *local_228 == 0)) {
    (**(code **)(local_228[1] + 0x30))();
  }
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  if (((local_1d0 != (longlong *)0x0) && (-1 < (int)*local_1d0)) &&
     (*local_1d0 = *local_1d0 + -1, *local_1d0 == 0)) {
    (**(code **)(local_1d0[1] + 0x30))();
  }
  if (((local_220 != (longlong *)0x0) && (-1 < (int)*local_220)) &&
     (*local_220 = *local_220 + -1, *local_220 == 0)) {
    (**(code **)(local_220[1] + 0x30))();
  }
  local_2b0 = lVar5;
  if ((-1 < (int)*local_218) && (*local_218 = *local_218 + -1, *local_218 == 0)) {
    (**(code **)(local_218[1] + 0x30))();
  }
  if ((-1 < (int)*local_210) && (*local_210 = *local_210 + -1, *local_210 == 0)) {
    (**(code **)(local_210[1] + 0x30))();
  }
  if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
    (**(code **)(plVar6[1] + 0x30))(plVar6);
  }
  if ((-1 < (int)*local_238) && (*local_238 = *local_238 + -1, *local_238 == 0)) {
    (**(code **)(local_238[1] + 0x30))();
  }
  if ((-1 < (int)*local_208) && (*local_208 = *local_208 + -1, *local_208 == 0)) {
    (**(code **)(local_208[1] + 0x30))();
  }
  if ((-1 < (int)*local_200) && (*local_200 = *local_200 + -1, *local_200 == 0)) {
    (**(code **)(local_200[1] + 0x30))();
  }
  if ((-1 < (int)*local_1c8) && (*local_1c8 = *local_1c8 + -1, *local_1c8 == 0)) {
    (**(code **)(local_1c8[1] + 0x30))();
  }
  FUN_18000ce40(param_1,&local_2b0);
  return 0;
}

